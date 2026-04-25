package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig // 新增 JWT 配置
}

type AppConfig struct {
	Name        string
	Mode        string
	Port        int
	UserPort    int `mapstructure:"user_port"`
	ProductPort int `mapstructure:"product_port"`
	OrderPort   int `mapstructure:"order_port"`
	NodeID      int64 `mapstructure:"node_id"`
}

type MySQLConfig struct {
	DSN          string
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	MaxOpenConns int `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int `mapstructure:"pool_size"`
}

type RabbitMQConfig struct {
	URL string
}
type JWTConfig struct {
	Secret        string
	AccessExpire  int `mapstructure:"access_expire"`
	RefreshExpire int `mapstructure:"refresh_expire"`
}

var GlobalConfig *Config

// LoadConfig 读取并解析 yaml 配置
func LoadConfig(path string) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	if err := validateConfig(GlobalConfig); err != nil {
		log.Fatalf("配置校验失败: %v", err)
	}
	log.Println("配置文件加载成功!")
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("配置为空")
	}

	if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
		return fmt.Errorf("app.port 非法: %d", cfg.App.Port)
	}

	if cfg.App.UserPort <= 0 || cfg.App.UserPort > 65535 {
		return fmt.Errorf("app.user_port 非法: %d", cfg.App.UserPort)
	}

	if cfg.App.ProductPort <= 0 || cfg.App.ProductPort > 65535 {
		return fmt.Errorf("app.product_port 非法: %d", cfg.App.ProductPort)
	}

	if cfg.App.OrderPort <= 0 || cfg.App.OrderPort > 65535 {
		return fmt.Errorf("app.order_port 非法: %d", cfg.App.OrderPort)
	}

	mode := strings.ToLower(strings.TrimSpace(cfg.App.Mode))
	if mode != "debug" && mode != "release" {
		return fmt.Errorf("app.mode 仅支持 debug/release, 当前: %s", cfg.App.Mode)
	}

	if cfg.App.NodeID < 0 || cfg.App.NodeID > 1023 {
		return fmt.Errorf("app.node_id 需在 [0,1023], 当前: %d", cfg.App.NodeID)
	}

	if strings.TrimSpace(cfg.MySQL.DSN) == "" {
		return errors.New("mysql.dsn 不能为空")
	}
	if cfg.MySQL.MaxIdleConns <= 0 {
		return fmt.Errorf("mysql.max_idle_conns 需 > 0, 当前: %d", cfg.MySQL.MaxIdleConns)
	}
	if cfg.MySQL.MaxOpenConns <= 0 {
		return fmt.Errorf("mysql.max_open_conns 需 > 0, 当前: %d", cfg.MySQL.MaxOpenConns)
	}

	if strings.TrimSpace(cfg.Redis.Addr) == "" {
		return errors.New("redis.addr 不能为空")
	}
	if cfg.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis.pool_size 需 > 0, 当前: %d", cfg.Redis.PoolSize)
	}

	if strings.TrimSpace(cfg.RabbitMQ.URL) == "" {
		return errors.New("rabbitmq.url 不能为空")
	}

	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return errors.New("jwt.secret 不能为空")
	}
	if cfg.JWT.AccessExpire <= 0 {
		return fmt.Errorf("jwt.access_expire 需 > 0, 当前: %d", cfg.JWT.AccessExpire)
	}
	if cfg.JWT.RefreshExpire <= 0 {
		return fmt.Errorf("jwt.refresh_expire 需 > 0, 当前: %d", cfg.JWT.RefreshExpire)
	}

	return nil
}
