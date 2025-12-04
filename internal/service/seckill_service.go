package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	radix "github.com/mediocregopher/radix/v3"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
)

const (
	redisSeckillPathKey    = "seckill:path:%d:%d" // userID, productID
	redisSeckillStockKey   = "seckill:stock:%d"   // productID
	redisSeckillSuccessKey = "seckill:succ:%d:%d" // userID, productID

	seckillQueue = "seckill_queue"
)

type SeckillMessage struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
}

type SeckillService struct {
	productRepo product.Repository
	redis       radix.Client
	mqConn      *amqp.Connection
	jwtCfg      *config.JWTConfig
}

func NewSeckillService(
	productRepo product.Repository,
	redis radix.Client,
	mqConn *amqp.Connection,
	jwtCfg *config.JWTConfig,
) *SeckillService {
	return &SeckillService{
		productRepo: productRepo,
		redis:       redis,
		mqConn:      mqConn,
		jwtCfg:      jwtCfg,
	}
}

// InitProductStock 将商品秒杀库存同步到 Redis
func (s *SeckillService) InitProductStock(ctx context.Context, p *product.Product) error {
	key := fmt.Sprintf(redisSeckillStockKey, p.ID)
	return s.redis.Do(radix.FlatCmd(nil, "SET", key, p.SeckillStock))
}

// GeneratePath 生成动态秒杀地址
func (s *SeckillService) GeneratePath(ctx context.Context, userID, productID int64) (string, error) {
	raw := fmt.Sprintf("u%d-p%d-%d-%s", userID, productID, time.Now().UnixNano(), s.jwtCfg.Secret)
	sum := md5.Sum([]byte(raw))
	path := hex.EncodeToString(sum[:])

	key := fmt.Sprintf(redisSeckillPathKey, userID, productID)
	err := s.redis.Do(radix.FlatCmd(nil, "SETEX", key, 300, path)) // 5 分钟有效
	return path, err
}

// Seckill 发起秒杀：校验 path、预减库存、写 MQ
func (s *SeckillService) Seckill(ctx context.Context, userID, productID int64, path string) error {
	// 1. 校验 path
	pathKey := fmt.Sprintf(redisSeckillPathKey, userID, productID)
	var stored string
	if err := s.redis.Do(radix.Cmd(&stored, "GET", pathKey)); err != nil {
		return err
	}
	if stored == "" || stored != path {
		return fmt.Errorf("invalid seckill path")
	}

	// 2. 判断是否已成功过
	succKey := fmt.Sprintf(redisSeckillSuccessKey, userID, productID)
	var succ int
	_ = s.redis.Do(radix.Cmd(&succ, "EXISTS", succKey))
	if succ == 1 {
		return fmt.Errorf("duplicate seckill")
	}

	// 3. 预减库存
	stockKey := fmt.Sprintf(redisSeckillStockKey, productID)
	var left int
	if err := s.redis.Do(radix.Cmd(&left, "DECR", stockKey)); err != nil {
		return err
	}
	if left < 0 {
		// 回滚
		_ = s.redis.Do(radix.Cmd(nil, "INCR", stockKey))
		return fmt.Errorf("stock empty")
	}

	// 4. 写 MQ
	ch, err := s.mqConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err = ch.QueueDeclare(seckillQueue, true, false, false, false, nil); err != nil {
		return err
	}

	body, err := json.Marshal(&SeckillMessage{
		UserID:    userID,
		ProductID: productID,
	})
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		"",
		seckillQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
