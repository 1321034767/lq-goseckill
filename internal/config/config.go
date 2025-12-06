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

// AuthConfig 鉴权/一致性哈希配置
type AuthConfig struct {
	// Nodes 为参与一致性哈希环的节点标识（可用节点名/IP:port）
	Nodes []string
	// HashReplicas 虚拟节点倍数，用于平衡分布
	HashReplicas int
	// TokenCacheTTLSeconds JWT 解析结果缓存时间（秒）
	TokenCacheTTLSeconds int
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
	Auth        AuthConfig
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
		Auth: AuthConfig{
			Nodes:                []string{"auth-node-1", "auth-node-2", "auth-node-3"},
			HashReplicas:         50,
			TokenCacheTTLSeconds: 600,
		},
		JWT: JWTConfig{
			Secret: "goseckill-secret",
		},
	}
}
