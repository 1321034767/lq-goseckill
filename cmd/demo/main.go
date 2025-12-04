package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/mq"
	"github.com/example/goseckill/internal/infra/redis"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

// 简单 demo：初始化一个商品并把库存同步到 Redis，用于手工测试秒杀流程
func main() {
	cfg := config.DefaultConfig()
	db := mysql.Init(&cfg.MySQL)
	redisClient := redis.Init(&cfg.Redis)
	mqConn := mq.Init(&cfg.RabbitMQ)

	productRepo := mysql.NewProductRepository(db)

	// 创建一个秒杀商品
	p := &product.Product{
		Name:         "测试秒杀商品",
		Description:  "这是一个用于秒杀测试的商品",
		Price:        1000,
		Stock:        100,
		SeckillStock: 10,
		StartTime:    time.Now().Add(-time.Hour),
		EndTime:      time.Now().Add(time.Hour),
		Status:       1,
	}
	if err := productRepo.Create(context.Background(), p); err != nil {
		log.Fatalf("create product failed: %v", err)
	}

	// 同步秒杀库存到 Redis
	seckillSvc := service.NewSeckillService(productRepo, redisClient, mqConn, &cfg.JWT)
	if err := seckillSvc.InitProductStock(context.Background(), p); err != nil {
		log.Fatalf("init redis stock failed: %v", err)
	}

	fmt.Printf("demo 初始化完成，商品 ID = %d，秒杀库存 = %d\n", p.ID, p.SeckillStock)
	fmt.Println("现在你可以：")
	fmt.Println("1) 启动 web 服务：go run ./cmd/web")
	fmt.Println("2) 启动 worker： go run ./cmd/seckill-worker")
	fmt.Println("3) 使用 HTTP 工具调用 /api/register、/api/login、/api/seckill/{id}/path 和 /api/seckill/{id}/{path} 进行秒杀测试")
}
