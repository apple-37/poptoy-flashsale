package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"poptoy-flashsale/app/order/internal/bootstrap"
	"poptoy-flashsale/app/order/internal/router"
	"poptoy-flashsale/app/order/internal/service"
	"poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/config"
	"poptoy-flashsale/pkg/idgen"
	"poptoy-flashsale/pkg/mq"
	"poptoy-flashsale/pkg/mysql"

	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "../../conf/config.yaml", "配置文件路径")
	flag.Parse()

	// 如果使用环境变量覆盖
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		*configPath = envPath
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatalf("获取配置文件绝对路径失败: %v", err)
	}

	// 1. 加载配置
	config.LoadConfig(absPath)

	// 2. 初始化核心中间件与数据库
	mysql.InitDB()
	bootstrap.InitStorage()
	cache.InitRedis()
	if err := idgen.InitSnowflake(config.GlobalConfig.App.NodeID); err != nil {
		log.Fatalf("Snowflake 初始化失败: %v", err)
	}
	mq.InitRabbitMQ()
	defer mq.Close()
	service.InitFSM()

	// 3. 启动 Order 模块的后台消费者 Worker (监听秒杀与死信队列)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bootstrap.InitWorkers(ctx)

	// 4. 设置 Gin 引擎
	gin.SetMode(config.GlobalConfig.App.Mode)
	r := gin.Default()

	// 5. 注册路由
	apiV1 := r.Group("/api/v1")
	{
		router.RegisterRoutes(apiV1) // 订单秒杀模块 (/orders/...)
	}

	// 6. 启动 Web 服务
	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.OrderPort)
	log.Printf("==== Order Service 启动成功，监听端口 %s ====\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}