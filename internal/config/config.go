package config

import "fmt"

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Host string
	Port int
}

func (s ServerConfig) Addr() string {
	if s.Host == "" {
		s.Host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	DSN string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr string
}

// RabbitMQConfig MQ 配置
type RabbitMQConfig struct {
	URL string
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string
}

// Config 应用总配置
type Config struct {
	Server      ServerConfig
	AdminServer ServerConfig
	MySQL       MySQLConfig
	Redis       RedisConfig
	RabbitMQ    RabbitMQConfig
	JWT         JWTConfig
}

// DefaultConfig 默认配置，方便快速跑起来
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		AdminServer: ServerConfig{
			Host: "0.0.0.0",
			Port: 8081,
		},
		MySQL: MySQLConfig{
			DSN: "goseckill:goseckill123@tcp(127.0.0.1:3306)/goseckill?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
		},
		RabbitMQ: RabbitMQConfig{
			URL: "amqp://guest:guest@127.0.0.1:5672/",
		},
		JWT: JWTConfig{
			Secret: "goseckill-secret",
		},
	}
}


