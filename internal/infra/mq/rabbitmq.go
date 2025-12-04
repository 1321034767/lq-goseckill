package mq

import (
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/goseckill/internal/config"
)

var (
	conn *amqp.Connection
	once sync.Once
)

// Init 初始化 RabbitMQ 连接
func Init(cfg *config.RabbitMQConfig) *amqp.Connection {
	once.Do(func() {
		c, err := amqp.Dial(cfg.URL)
		if err != nil {
			log.Fatalf("failed to connect rabbitmq: %v", err)
		}
		conn = c
	})
	return conn
}

// Conn 获取 MQ 连接
func Conn() *amqp.Connection {
	return conn
}


