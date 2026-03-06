package config

import (
	"log"

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
	Name string
	Mode string
	Port int
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
	log.Println("配置文件加载成功!")
}