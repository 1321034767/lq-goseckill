package main

import (
	"context"
	"fmt"
	"log"
	"time"

	radix "github.com/mediocregopher/radix/v3"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/redis"
	"github.com/example/goseckill/internal/repository/mysql"
)

const (
	redisSeckillStockKey = "seckill:stock:%d" // productID
	checkInterval        = 5 * time.Minute    // 每5分钟检查一次
)

func main() {
	cfg := config.DefaultConfig()

	db := mysql.Init(&cfg.MySQL)
	redisClient := redis.Init(&cfg.Redis)
	productRepo := mysql.NewProductRepository(db)

	log.Println("库存一致性检查服务启动...")
	log.Printf("检查间隔: %v", checkInterval)

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// 立即执行一次
	checkAndSync(context.Background(), productRepo, redisClient)

	// 定时执行
	for range ticker.C {
		checkAndSync(context.Background(), productRepo, redisClient)
	}
}

func checkAndSync(ctx context.Context, productRepo product.Repository, redisClient radix.Client) {
	log.Println("开始检查库存一致性...")

	// 获取所有商品
	products, err := productRepo.ListAll(ctx)
	if err != nil {
		log.Printf("获取商品列表失败: %v", err)
		return
	}

	inconsistentCount := 0
	syncedCount := 0

	// 只检查秒杀中的商品（Status=2）
	for _, p := range products {
		if p.Status != 2 {
			continue // 跳过非秒杀商品
		}

		if err := checkProductStock(ctx, p.ID, p.SeckillStock, redisClient); err != nil {
			log.Printf("检查商品 %d 失败: %v", p.ID, err)
			continue
		}

		stockKey := fmt.Sprintf(redisSeckillStockKey, p.ID)
		var redisStock int
		if err := redisClient.Do(radix.Cmd(&redisStock, "GET", stockKey)); err != nil {
			// Redis中没有，需要同步
			if err := syncStockToRedis(ctx, p.ID, p.SeckillStock, redisClient); err == nil {
				syncedCount++
			}
			continue
		}

		if int64(redisStock) != p.SeckillStock {
			inconsistentCount++
			log.Printf("⚠️  商品 %d (%s): 库存不一致 - MySQL: %d, Redis: %d", p.ID, p.Name, p.SeckillStock, redisStock)
			// 以MySQL为准，同步到Redis
			if err := syncStockToRedis(ctx, p.ID, p.SeckillStock, redisClient); err == nil {
				syncedCount++
				log.Printf("✅ 商品 %d: 已修复库存不一致", p.ID)
			}
		}
	}

	log.Printf("库存一致性检查完成 - 发现不一致: %d 个, 已修复: %d 个", inconsistentCount, syncedCount)
}

func syncStockToRedis(ctx context.Context, productID int64, stock int64, redisClient radix.Client) error {
	stockKey := fmt.Sprintf(redisSeckillStockKey, productID)
	if err := redisClient.Do(radix.FlatCmd(nil, "SET", stockKey, stock)); err != nil {
		return fmt.Errorf("同步库存到Redis失败: %v", err)
	}
	return nil
}
