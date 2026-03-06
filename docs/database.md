### 📁 3. docs/database.md (数据库设计)

```sql
-- 数据库设计与建表语句 (MySQL 8.0)
-- 引擎：InnoDB
-- 字符集：utf8mb4 (支持Emoji，国际化标准)

CREATE DATABASE IF NOT EXISTS `blindbox_mall` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `blindbox_mall`;

-- 1. 用户表
-- 设计说明: password_hash使用Bcrypt，长度需设为128。主键自增以优化聚簇索引。
CREATE TABLE `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `username` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户名',
    `password_hash` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码哈希',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 2. 商品热数据表 (垂直分表)
-- 设计说明: 剔除长文本，单行数据极小，Buffer Pool缓存命中率极高，适合高频列表展示和锁竞争。
CREATE TABLE `product_hot` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品标题',
    `price` DECIMAL(10,2) NOT NULL DEFAULT '0.00' COMMENT '价格',
    `stock` INT NOT NULL DEFAULT 0 COMMENT '真实剩余库存',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1:上架 0:下架',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_status_created` (`status`, `created_at`) -- 联合索引，加速列表查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品高频热表';

-- 3. 商品冷数据表 (垂直分表)
-- 设计说明: 存放 description 长文本，仅在查询详情时关联。主键直接使用 product_id，不单独设自增键。
CREATE TABLE `product_cold` (
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '关联热表ID',
    `description` TEXT NOT NULL COMMENT '长文本描述',
    `images_json` JSON NOT NULL COMMENT '图片列表JSON',
    PRIMARY KEY (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品低频冷表';

-- 4. 订单表
-- 设计说明: status做状态机(0:初始化, 1:待支付, 2:已支付, 3:已取消)。
-- 索引遵循最左匹配原则：业务中高频查询为“某用户按时间倒序查看订单”和“根据唯一订单号查询”。
CREATE TABLE `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '全局唯一订单号(雪花算法)',
    `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '购买人',
    `product_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品ID',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '订单状态',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_status_time` (`user_id`, `status`, `created_at`) -- 最左匹配核心索引
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
```