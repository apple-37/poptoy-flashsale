package cache

import (
	"context"
	"log"

	"poptoy-flashsale/pkg/config"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

// InitRedis 初始化 Redis 连接池
func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.GlobalConfig.Redis.Addr,
		Password: config.GlobalConfig.Redis.Password,
		DB:       config.GlobalConfig.Redis.DB,
		PoolSize: config.GlobalConfig.Redis.PoolSize,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	log.Println("Redis 连接成功!")

	// 初始化布隆过滤器
	InitBloomFilters()
}