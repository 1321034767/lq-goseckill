package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	radix "github.com/mediocregopher/radix/v3"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/mq"
	"github.com/example/goseckill/internal/infra/redis"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

func init() {
	// 初始化监控
	_ = service.GetMonitor()
}

const (
	seckillQueue             = "seckill_queue"
	redisSeckillStockKey     = "seckill:stock:%d"   // productID
	redisSeckillSuccessKey   = "seckill:succ:%d:%d" // userID, productID
	successMarkExpireSeconds = 86400                // 24小时有效期
)

func main() {
	cfg := config.DefaultConfig()

	db := mysql.Init(&cfg.MySQL)
	mqConn := mq.Init(&cfg.RabbitMQ)
	redisClient := redis.Init(&cfg.Redis)

	productRepo := mysql.NewProductRepository(db)
	orderRepo := mysql.NewOrderRepository(db)
	userRepo := mysql.NewUserRepository(db)
	accountSvc := service.NewAccountService(db, productRepo, orderRepo, userRepo)
	activityRepo := mysql.NewSeckillActivityRepository(db)
	activitySvc := service.NewSeckillActivityService(activityRepo, productRepo)

	ch, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	if _, err = ch.QueueDeclare(seckillQueue, true, false, false, false, nil); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// 改为手动确认模式（auto-ack=false）
	msgs, err := ch.Consume(seckillQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	log.Println("seckill worker started, waiting for messages...")

	for d := range msgs {
		var m service.SeckillMessage
		if err := json.Unmarshal(d.Body, &m); err != nil {
			log.Printf("invalid message: %v", err)
			// 消息格式错误，拒绝并丢弃
			_ = d.Nack(false, false)
			continue
		}
		handleMessage(context.Background(), productRepo, activitySvc, accountSvc, redisClient, &m, d)
	}
}

func handleMessage(ctx context.Context, productRepo product.Repository, activitySvc *service.SeckillActivityService, accountSvc *service.AccountService, redisClient radix.Client, m *service.SeckillMessage, d amqp.Delivery) {
	// 记录是否需要回滚Redis库存
	needRollback := false
	stockKey := fmt.Sprintf(redisSeckillStockKey, m.ProductID)

	// 使用defer确保失败时回滚库存
	defer func() {
		if needRollback {
			// 回滚Redis库存
			if err := redisClient.Do(radix.Cmd(nil, "INCR", stockKey)); err != nil {
				log.Printf("failed to rollback redis stock: %v", err)
			} else {
				log.Printf("rolled back redis stock for product %d", m.ProductID)
			}
		}
	}()

	p, err := productRepo.GetByID(ctx, m.ProductID)
	if err != nil {
		log.Printf("get product failed: %v", err)
		service.GetMonitor().RecordDBError()
		service.GetMonitor().RecordWorkerFailed()
		needRollback = true
		// 拒绝消息并重新入队
		_ = d.Nack(false, true)
		return
	}
	if p.SeckillStock <= 0 {
		log.Printf("product %d stock empty", p.ID)
		needRollback = true
		// 拒绝消息并重新入队
		_ = d.Nack(false, true)
		return
	}

	// 先扣减 MySQL 中的秒杀库存
	p.SeckillStock--
	if err := productRepo.Update(ctx, p); err != nil {
		log.Printf("update product stock failed: %v", err)
		needRollback = true
		// 拒绝消息并重新入队
		_ = d.Nack(false, true)
		return
	}

	// 计算本次应扣的秒杀价：默认原价，若有进行中活动且折扣合法则按折扣价
	priceToCharge := p.Price
	if activitySvc != nil {
		act, err := activitySvc.GetActivityByProduct(ctx, m.ProductID)
		if err == nil && act != nil {
			now := time.Now()
			if act.Status == 1 && now.After(act.StartTime) && now.Before(act.EndTime) && act.Discount > 0 && act.Discount <= 1 {
				priceToCharge = int64(math.Round(float64(p.Price) * act.Discount))
			}
		}
	}

	// 使用账户服务完成扣费 + 订单创建 + 流水记录
	o, err := accountSvc.SeckillCharge(ctx, m.UserID, m.ProductID, priceToCharge)
	if err != nil {
		log.Printf("seckill charge failed: %v", err)
		needRollback = true
		// 回滚 MySQL 库存
		p.SeckillStock++
		_ = productRepo.Update(ctx, p)
		// 拒绝消息并重新入队
		_ = d.Nack(false, true)
		return
	}

	// 递增用户对该商品的秒杀成功次数（用于每人限购统计）
	succKey := fmt.Sprintf(redisSeckillSuccessKey, m.UserID, m.ProductID)
	var newCount int
	if err := redisClient.Do(radix.Cmd(&newCount, "INCR", succKey)); err != nil {
		log.Printf("failed to increase seckill success count: %v", err)
	} else {
		// 首次成功时设置过期时间，避免长期占用Redis
		if newCount == 1 {
			if err := redisClient.Do(radix.Cmd(nil, "EXPIRE", succKey, strconv.Itoa(successMarkExpireSeconds))); err != nil {
				log.Printf("failed to set expire for success count key: %v", err)
			}
		}
		log.Printf("increase seckill success count: user=%d product=%d count=%d", m.UserID, m.ProductID, newCount)
	}

	log.Printf("create order success, order_id=%d user=%d product=%d", o.ID, o.UserID, m.ProductID)
	service.GetMonitor().RecordWorkerProcessed()

	// 处理成功，确认消息
	if err := d.Ack(false); err != nil {
		log.Printf("failed to ack message: %v", err)
	}
}
