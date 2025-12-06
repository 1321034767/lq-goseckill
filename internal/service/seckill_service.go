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
	"github.com/example/goseckill/internal/datamodels/seckill_activity"
)

const (
	redisSeckillPathKey    = "seckill:path:%d:%d"           // userID, productID
	redisSeckillStockKey   = "seckill:stock:%d"             // productID
	redisSeckillSuccessKey = "seckill:succ:%d:%d"           // userID, productID (成功标记，供结果查询/幂等使用)
	redisSeckillLimitKey   = "seckill:limit:%d:%d:%d"       // userID, productID, activityID（每个活动单独计数）

	seckillQueue = "seckill_queue"
)

type SeckillMessage struct {
	UserID    int64 `json:"user_id"`
	ProductID int64 `json:"product_id"`
}

type SeckillService struct {
	productRepo  product.Repository
	activityRepo seckill_activity.Repository
	redis        radix.Client
	mqConn       *amqp.Connection
	jwtCfg       *config.JWTConfig
}

func NewSeckillService(
	productRepo product.Repository,
	activityRepo seckill_activity.Repository,
	redis radix.Client,
	mqConn *amqp.Connection,
	jwtCfg *config.JWTConfig,
) *SeckillService {
	return &SeckillService{
		productRepo:  productRepo,
		activityRepo: activityRepo,
		redis:        redis,
		mqConn:       mqConn,
		jwtCfg:       jwtCfg,
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
	GetMonitor().RecordSeckillRequest()
	// 0. 获取商品信息并校验时间和状态
	p, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return fmt.Errorf("product not found: %v", err)
	}
	
	// 校验商品状态（必须是秒杀中）
	if p.Status != 2 {
		return fmt.Errorf("商品当前不在秒杀状态")
	}
	
	// 校验秒杀时间
	now := time.Now()
	if now.Before(p.StartTime) {
		GetMonitor().RecordSeckillError()
		return fmt.Errorf("秒杀尚未开始")
	}
	if now.After(p.EndTime) {
		GetMonitor().RecordSeckillError()
		return fmt.Errorf("秒杀已结束")
	}
	
	// 1. 校验 path
	pathKey := fmt.Sprintf(redisSeckillPathKey, userID, productID)
	var stored string
	if err := s.redis.Do(radix.Cmd(&stored, "GET", pathKey)); err != nil {
		return err
	}
	if stored == "" || stored != path {
		return fmt.Errorf("秒杀地址无效或已过期")
	}

	// 2. 校验当前用户是否超过“每人限购”次数（基于活动配置，使用 Redis 计数）
	limit := int64(1)
	var activeActID int64
	if s.activityRepo != nil {
		activities, err := s.activityRepo.GetActivitiesByProduct(ctx, productID)
		if err == nil && len(activities) > 0 {
			now := time.Now()
			for _, act := range activities {
				if act.Status == 1 && now.After(act.StartTime) && now.Before(act.EndTime) {
					activeActID = act.ID
					if act.LimitPerUser > 0 {
						limit = act.LimitPerUser
					}
					break
				}
			}
		}
	}

	// 如果没找到当前正在进行的活动，说明配置有问题或活动已结束
	if activeActID == 0 {
		return fmt.Errorf("当前没有进行中的秒杀活动")
	}

	limitKey := fmt.Sprintf(redisSeckillLimitKey, userID, productID, activeActID)
	var used int
	// INCR 原子增加已抢购次数
	if err := s.redis.Do(radix.Cmd(&used, "INCR", limitKey)); err != nil {
		return err
	}
	// 第一次成功时设置过期时间（这里简单设置 24 小时，可按需调整）
	if used == 1 {
		_ = s.redis.Do(radix.Cmd(nil, "EXPIRE", limitKey, "86400"))
	}
	if int64(used) > limit {
		// 超出限购，回滚计数并返回错误
		_ = s.redis.Do(radix.Cmd(nil, "DECR", limitKey))
		return fmt.Errorf("超过每人限购数量，无法继续秒杀")
	}

	// 3. 预减库存
	stockKey := fmt.Sprintf(redisSeckillStockKey, productID)
	var left int
	if err := s.redis.Do(radix.Cmd(&left, "DECR", stockKey)); err != nil {
		GetMonitor().RecordRedisError()
		return err
	}
	if left < 0 {
		// 回滚
		_ = s.redis.Do(radix.Cmd(nil, "INCR", stockKey))
		return fmt.Errorf("秒杀库存不足")
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

	err = ch.PublishWithContext(
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
	if err != nil {
		GetMonitor().RecordMQError()
		return err
	}
	GetMonitor().RecordSeckillSuccess()
	return nil
}
