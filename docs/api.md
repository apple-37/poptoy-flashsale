**Base URL**: `http://localhost:8080/api/v1`
**Authorization**: Bearer `<JWT_TOKEN>`

---
### 1. User Service
#### 1.1 用户注册
*   **POST** `/users/register`
*   **Request Body**: `{"username": "testuser", "password": "pwd"}`
*   **Responses**:
    *   `201 Created`: `{"code": 20100, "msg": "注册成功"}`
    *   `400 Bad Request`: `{"code": 40001, "msg": "用户名已存在"}`
    *   `500 Internal Server Error`: `{"code": 50000, "msg": "服务器内部错误"}`

#### 1.2 用户登录
*   **POST** `/users/login`
*   **Request Body**: `{"username": "testuser", "password": "pwd"}`
*   **Responses**:
    *   `200 OK`: `{"code": 20000, "msg": "登录成功", "data": {"token": "eyJhb..."}}`
    *   `401 Unauthorized`: `{"code": 40101, "msg": "用户名或密码错误"}`

---
### 2. Product Service
#### 2.1 获取商品列表
*   **GET** `/products?page=1&size=10`
*   **Responses**:
    *   `200 OK`: 返回 `product_hot` 列表数据。

---
### 3. Order Service
#### 3.1 发起秒杀 (异步)
*   **POST** `/orders/flash-buy`
*   **Headers**: `Authorization: Bearer <token>`
*   **Request Body**: `{"product_id": 101}`
*   **Responses**:
    *   `202 Accepted`: 
        ```json
        {"code": 20200, "msg": "抢购任务已受理，请监听推送", "data": null}
        ```
    *   `400 Bad Request`: `{"code": 40010, "msg": "商品已售罄"}`
    *   `401 Unauthorized`: `{"code": 40100, "msg": "Token缺失或无效"}`

#### 3.2 监听秒杀结果 (SSE 长连接)
*   **GET** `/orders/result/stream`
*   **Headers**: `Authorization: Bearer <token>`
*   **描述**: 前端通过 `EventSource` 建立连接。后端阻塞监听 Redis Pub/Sub，一旦 Worker 创建订单完毕，推送如下格式数据：
*   **SSE Push Data** (Text/Event-Stream):
    ```text
    event: order_success
    data: {"order_no": "ORD2026...", "status": "pending"}
    ```
```