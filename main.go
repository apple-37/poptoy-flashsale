package main

import (
	"context"
	"fmt"
	"log"

	"poptoy-flashsale/app/order"
	"poptoy-flashsale/app/product"
	"poptoy-flashsale/app/user"
	"poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/config"
	"poptoy-flashsale/pkg/mq"
	"poptoy-flashsale/pkg/mysql"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 加载配置
	config.LoadConfig("./conf/config.yaml")

	// 2. 初始化核心中间件与数据库
	mysql.InitDB()
	cache.InitRedis()
	mq.InitRabbitMQ()
	defer mq.Close()

	// 3. 启动 Order 模块的后台消费者 Worker (监听秒杀与死信队列)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	order.InitWorkers(ctx)

	// 4. 设置 Gin 引擎
	gin.SetMode(config.GlobalConfig.App.Mode)
	r := gin.Default()

	// 5. 注册路由 (统一调用各个微服务模块对外的 RegisterRoutes)
	apiV1 := r.Group("/api/v1")
	{
		user.RegisterRoutes(apiV1)    // 用户模块 (/users/...)
		product.RegisterRoutes(apiV1) // 商品模块 (/products/...)
		order.RegisterRoutes(apiV1)   // 订单秒杀模块 (/orders/...)
	}

	// 6. 启动 Web 服务
	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.Port)
	log.Printf("==== PopToy FlashSale 系统启动成功，监听端口 %s ====\n", addr)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}