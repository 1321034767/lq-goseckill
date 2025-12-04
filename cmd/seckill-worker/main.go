package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/order"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/mq"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

const seckillQueue = "seckill_queue"

func main() {
	cfg := config.DefaultConfig()

	db := mysql.Init(&cfg.MySQL)
	mqConn := mq.Init(&cfg.RabbitMQ)

	productRepo := mysql.NewProductRepository(db)
	orderRepo := mysql.NewOrderRepository(db)

	ch, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	if _, err = ch.QueueDeclare(seckillQueue, true, false, false, false, nil); err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(seckillQueue, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	log.Println("seckill worker started, waiting for messages...")

	for d := range msgs {
		var m service.SeckillMessage
		if err := json.Unmarshal(d.Body, &m); err != nil {
			log.Printf("invalid message: %v", err)
			continue
		}
		handleMessage(context.Background(), productRepo, orderRepo, &m)
	}
}

func handleMessage(ctx context.Context, productRepo product.Repository, orderRepo order.Repository, m *service.SeckillMessage) {
	p, err := productRepo.GetByID(ctx, m.ProductID)
	if err != nil {
		log.Printf("get product failed: %v", err)
		return
	}
	if p.SeckillStock <= 0 {
		log.Printf("product %d stock empty", p.ID)
		return
	}
	p.SeckillStock--
	if err := productRepo.Update(ctx, p); err != nil {
		log.Printf("update product stock failed: %v", err)
		return
	}

	o := &order.Order{
		UserID:    m.UserID,
		ProductID: m.ProductID,
		Price:     p.Price,
		Status:    0,
	}
	if err := orderRepo.Create(ctx, o); err != nil {
		log.Printf("create order failed: %v", err)
		return
	}
	log.Printf("create order success, order_id=%d user=%d product=%d", o.ID, o.UserID, o.ProductID)
}
