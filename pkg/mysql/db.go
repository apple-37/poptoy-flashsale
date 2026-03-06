package mysql

import (
	"log"
	"time"

	"poptoy-flashsale/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化 MySQL 连接池
func InitDB() {
	var err error
	
	// 根据运行模式设置日志级别
	logLevel := logger.Info
	if config.GlobalConfig.App.Mode == "release" {
		logLevel = logger.Silent
	}

	DB, err = gorm.Open(mysql.Open(config.GlobalConfig.MySQL.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("MySQL 连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取底层 sql.DB 失败: %v", err)
	}

	// 生产级连接池配置
	sqlDB.SetMaxIdleConns(config.GlobalConfig.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.GlobalConfig.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("MySQL 连接成功, 连接池已配置!")
}