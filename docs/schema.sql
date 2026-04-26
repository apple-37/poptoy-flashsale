-- ============================================
-- PopToy 秒杀系统数据库建表脚本
-- 数据库: poptoy_flashsale
-- 创建时间: 2026-04-25
-- ============================================

CREATE DATABASE IF NOT EXISTS poptoy_flashsale_v2 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'PopToy 秒杀系统数据库';

USE poptoy_flashsale_v2;

-- ============================================
-- 1. 用户表 (users)
-- 存储用户基本信息
-- 状态流转: 0(已注册) -> 1(活跃) -> 2(禁用) -> 3(已删除)
-- ============================================
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
    `password_hash` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码哈希',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=已注册, 1=活跃, 2=禁用, 3=已删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 2. 商品热数据表 (product_hot)
-- 存储商品热数据，高频查询、参与锁竞争
-- 状态流转: 0(待审核) -> 1(上架) -> 2(下架) -> 3(售罄) -> 4(预售)
-- ============================================
CREATE TABLE IF NOT EXISTS `product_hot` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品标题',
    `price` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '商品价格',
    `stock` INT NOT NULL DEFAULT 0 COMMENT '库存数量',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=待审核, 1=上架, 2=下架, 3=售罄, 4=预售',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品热数据表';

-- ============================================
-- 3. 商品冷数据表 (product_cold)
-- 存储商品冷数据，长文本，低频查询
-- 与 product_hot 一一对应，通过 product_id 关联
-- ============================================
CREATE TABLE IF NOT EXISTS `product_cold` (
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID(对应热表ID)',
    `description` TEXT NOT NULL COMMENT '商品描述',
    `images_json` JSON NOT NULL COMMENT '商品图片JSON数组',
    PRIMARY KEY (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品冷数据表';

-- ============================================
-- 4. 订单表 (orders)
-- 存储秒杀订单信息
-- 状态流转: 0(Init) -> 1(Pending) -> 2(Paid) -> 3(Cancelled)
-- ============================================
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `order_no` VARCHAR(64) NOT NULL COMMENT '订单号',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=Init, 1=Pending, 2=Paid, 3=Cancelled',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_status_time` (`user_id`, `status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- ============================================
-- 5. 本地事务消息表 (order_outbox_messages)
-- 用于保障 MQ 发送失败后的补偿重试
-- 实现可靠消息投递
-- ============================================
CREATE TABLE IF NOT EXISTS `order_outbox_messages` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
    `message_type` VARCHAR(64) NOT NULL COMMENT '消息类型: create_order_task, delay_cancel_task',
    `biz_key` VARCHAR(128) NOT NULL COMMENT '业务键: flash:{orderNo}, delay:{orderNo}',
    `payload` LONGTEXT NOT NULL COMMENT '消息体JSON',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=Pending, 1=Sent, 2=Failed',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    `next_retry_at` DATETIME NOT NULL COMMENT '下次重试时间',
    `last_error` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最后错误信息',
    `sent_at` DATETIME DEFAULT NULL COMMENT '发送成功时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_outbox_biz_key` (`biz_key`),
    KEY `idx_outbox_type_status_next` (`message_type`, `status`, `next_retry_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='本地事务消息表';

-- ============================================
-- 索引说明
-- ============================================
-- users: username 唯一索引，支持快速查找
-- orders: (user_id, status, created_at) 复合索引，支持用户订单查询和状态筛选
-- order_outbox_messages: (message_type, status, next_retry_at) 复合索引，支持补偿扫描