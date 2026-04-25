# PopToy 秒杀系统

## 项目结构

```
poptoy-flashsale/
├── app/                          # 业务模块
│   ├── user/                     # 用户服务
│   │   ├── cmd/                  # 服务入口
│   │   └── internal/             # 内部代码
│   │       ├── controller/       # 控制器
│   │       ├── model/            # 数据模型
│   │       ├── repository/       # 数据访问层
│   │       ├── router/           # 路由注册
│   │       └── service/          # 业务逻辑
│   ├── product/                  # 商品服务
│   │   ├── cmd/                  # 服务入口
│   │   └── internal/             # 内部代码
│   │       ├── controller/       # 控制器
│   │       ├── model/            # 数据模型
│   │       ├── repository/       # 数据访问层
│   │       ├── router/           # 路由注册
│   │       └── service/          # 业务逻辑
│   └── order/                    # 订单服务
│       ├── cmd/                  # 服务入口
│       └── internal/             # 内部代码
│           ├── bootstrap/        # 初始化引导
│           ├── cache/            # 缓存处理
│           ├── controller/       # 控制器
│           ├── model/            # 数据模型
│           ├── mq/               # 消息队列
│           ├── repository/       # 数据访问层
│           ├── router/           # 路由注册
│           └── service/          # 业务逻辑
├── pkg/                          # 公共工具库
│   ├── cache/                   # Redis 缓存
│   ├── config/                   # 配置管理
│   ├── e/                        # 错误码
│   ├── fsm/                      # 有限状态机
│   ├── idgen/                    # ID 生成器
│   ├── jwt/                      # JWT 认证
│   ├── middleware/               # 中间件
│   ├── mq/                       # 消息队列
│   ├── mysql/                    # 数据库连接
│   └── response/                 # 统一响应
├── conf/                         # 配置文件
├── benchmark/                     # 基准测试
├── go.mod                        # Go 模块文件
└── README.md                     # 项目说明
```

## 服务说明

| 服务            | 端口 | 说明                                   |
| --------------- | ---- | -------------------------------------- |
| User Service    | 8081 | 用户管理服务，处理用户注册、登录等     |
| Product Service | 8082 | 商品管理服务，处理商品查询、状态管理等 |
| Order Service   | 8083 | 订单秒杀服务，处理秒杀请求、订单创建等 |

## 环境要求

- Go 1.18+
- MySQL 5.7+
- Redis 6.0+
- RabbitMQ 3.8+

## 配置文件

1. 复制 `conf/config.example.yaml` 为 `conf/config.yaml`
2. 修改配置文件中的数据库连接、Redis 连接、RabbitMQ 连接等信息
3. 确保各个服务的端口配置正确

## 启动方式

### 1. 启动用户服务

```bash
cd app/user/cmd
go run main.go -config ../../conf/config.yaml
```

### 2. 启动商品服务

```bash
cd app/product/cmd
go run main.go -config ../../conf/config.yaml
```

### 3. 启动订单服务

```bash
cd app/order/cmd
go run main.go -config ../../conf/config.yaml
```

### 4. 配置参数说明

```bash
-config string
    配置文件路径 (默认 "../../conf/config.yaml")
```

也可以使用环境变量 `CONFIG_PATH` 指定配置文件路径：

```bash
export CONFIG_PATH=/path/to/config.yaml
cd app/user/cmd && go run main.go
```

## API 接口

### 用户模块

- POST /api/v1/users/register - 用户注册
- POST /api/v1/users/login - 用户登录
- POST /api/v1/users/logout - 用户登出（需认证）
- POST /api/v1/users/refresh - 刷新 Token

### 商品模块

- GET /api/v1/products - 获取商品列表
- GET /api/v1/products/:id - 获取商品详情

### 订单模块

- POST /api/v1/orders/flash-buy - 秒杀下单（需认证）
- GET /api/v1/orders/result/stream - 订单状态 SSE 通知（需认证）

## 技术栈

- Go 1.18+
- Gin Web 框架
- MySQL 数据库
- Redis 缓存
- RabbitMQ 消息队列
- JWT 认证
- 雪花算法 ID 生成
- 有限状态机 (FSM) 状态管理

## 核心功能

1. **秒杀功能**：高并发秒杀，使用 Redis 预减库存，RabbitMQ 异步处理
2. **用户管理**：注册、登录、JWT 认证
3. **商品管理**：商品列表、详情、状态管理
4. **订单管理**：订单创建、状态流转、超时取消
5. **状态管理**：基于有限状态机的状态管理（用户、商品、订单）

## 注意事项

1. 确保 MySQL、Redis、RabbitMQ 服务正常运行
2. 确保配置文件中的连接信息正确
3. 启动服务时，建议先启动用户服务和商品服务，再启动订单服务
4. 秒杀活动期间，建议增加 Redis 和 RabbitMQ 的资源配置
