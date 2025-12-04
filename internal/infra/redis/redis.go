package redis

import (
	"log"
	"sync"

	radix "github.com/mediocregopher/radix/v3"

	"github.com/example/goseckill/internal/config"
)

var (
	client radix.Client
	once   sync.Once
)

// Init 初始化 Redis 连接池
func Init(cfg *config.RedisConfig) radix.Client {
	once.Do(func() {
		pool, err := radix.NewPool("tcp", cfg.Addr, 10)
		if err != nil {
			log.Fatalf("failed to connect redis: %v", err)
		}
		client = pool
	})
	return client
}

// Client 获取 Redis 客户端
func Client() radix.Client {
	return client
}


