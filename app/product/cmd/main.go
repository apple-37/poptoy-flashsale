package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"poptoy-flashsale/app/product/internal/router"
	"poptoy-flashsale/pkg/cache"
	"poptoy-flashsale/pkg/config"
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
	cache.InitRedis()

	// 3. 设置 Gin 引擎
	gin.SetMode(config.GlobalConfig.App.Mode)
	r := gin.Default()

	// 4. 注册路由
	apiV1 := r.Group("/api/v1")
	{
		router.RegisterRoutes(apiV1) // 商品模块 (/products/...)
	}

	// 5. 启动 Web 服务
	addr := fmt.Sprintf(":%d", config.GlobalConfig.App.ProductPort)
	log.Printf("==== Product Service 启动成功，监听端口 %s ====\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}