
# 🚀 PopToy-FlashSale | 在线盲盒与潮玩秒杀系统

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Gin](https://img.shields.io/badge/Gin-v1.9-00ADD8?style=flat&logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=flat&logo=redis)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.12-FF6600?style=flat&logo=rabbitmq)
![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat&logo=mysql)

> 一个基于 Go 语言实现的、高并发场景下的电商微服务后端系统。涵盖了秒杀预扣减、流量削峰、延迟队列超时控制、服务端实时推送 (SSE) 等核心技术。

## 📖 项目简介
本项目旨在解决电商大促场景下“高并发读写”、“超卖控制”与“分布式事务最终一致性”等核心痛点。
系统采用 **基于 Monorepo 的轻量级微服务架构**，核心业务包括 **限量潮玩秒杀** 与 **盲盒在线抽赏**。不同于传统的 CRUD 项目，本项目深度应用了 Redis Lua 脚本原子性、RabbitMQ 死信队列 (DLX) 以及有限状态机 (FSM) 设计模式。

## 🌟 核心特性 (Key Features)

*   **⚡ 极致的高并发防超卖**
    *   利用 **Redis + Lua 脚本** 实现库存预扣减与“一人一单”校验，将流量拦截在缓存层，保护下游数据库。
*   **🌊 流量削峰与异步解耦**
    *   引入 **RabbitMQ** 将秒杀请求异步化，通过 Worker 协程池平滑消费，防止数据库被瞬时流量击穿。
*   **⏰ 订单超时自动取消 (DLX)**
    *   基于 **RabbitMQ 死信队列 (Dead Letter Exchange)** 实现 15 分钟订单超时自动关闭与库存回滚，替代低效的定时轮询任务。
*   **📡 服务端实时推送 (SSE)**
    *   摒弃传统的客户端轮询 (Polling)，采用 **Server-Sent Events (SSE)** 结合 Redis Pub/Sub，实现秒杀结果的毫秒级实时推送。
*   **🛡️ 工业级安全与设计**
    *   **安全**：实现 **双 Token (Access/Refresh)** 机制，并针对登录接口采用恒定时间比对算法，防御 **时间侧信道攻击 (Timing Attack)**。
    *   **设计**：订单流转采用 **有限状态机 (FSM)** 模式，彻底杜绝并发下的状态混乱；数据库采用 **冷热数据分离 (Vertical Partitioning)** 设计。

## 🛠️ 技术栈 (Tech Stack)

*   **语言**: Golang 1.21+
*   **Web 框架**: Gin (RESTful API)
*   **数据库**: MySQL 8.0 (GORM v2, InnoDB, 垂直分表)
*   **缓存**: Redis 7.0 (Go-Redis v9, Lua Script, Pub/Sub)
*   **消息队列**: RabbitMQ (AMQP 0-9-1, Topic Exchange, DLX)
*   **配置管理**: Viper (支持 YAML/Env)
*   **鉴权**: JWT (Golang-JWT v5)
*   **环境**: WSL2 (Ubuntu) + Docker

## 📂 目录结构 (Project Layout)

```text
poptoy-flashsale/
├── app/                  # 微服务应用源码
│   ├── gateway/          # API 网关与路由分发
│   ├── user/             # 用户服务 (JWT, 安全登录)
│   ├── product/          # 商品服务 (冷热分离, 深分页优化)
│   └── order/            # 订单服务 (秒杀核心, FSM, MQ Worker)
├── pkg/                  # 公共基础设施层
│   ├── cache/            # Redis 封装
│   ├── mq/               # RabbitMQ 生产者/消费者封装
│   ├── fsm/              # 通用有限状态机实现
│   └── ...
├── conf/                 # 配置文件
├── docs/                 # 项目文档 (SRS, SDD, SQL)
└── main.go               # 系统启动入口
```

“针对高并发场景编写了 Go 协程压测脚本，在 1000 并发用户抢购 50 库存的测试中，系统 QPS 峰值达到 22,000+。通过 Redis Lua 原子与 RabbitMQ 削峰，最终 MySQL 落盘订单严格为 50 单，实现了 0 超卖与 0 宕机。”