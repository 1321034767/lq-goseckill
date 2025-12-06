package seckill_activity

import (
	"context"
	"time"
)

// SeckillActivity 秒杀活动模型
type SeckillActivity struct {
	ID          int64     `gorm:"primaryKey"`
	Name        string    `gorm:"size:128;not null"`        // 活动名称
	Description string    `gorm:"size:512"`                // 活动描述
	StartTime   time.Time `gorm:"index"`                   // 开始时间
	EndTime     time.Time `gorm:"index"`                   // 结束时间
	Discount    float64   `gorm:"type:decimal(5,2);not null"` // 折扣（0.1-1.0，如0.8表示8折）
	LimitPerUser int64   `gorm:"default:1"`                // 每人限购数量，默认1
	Status      int       `gorm:"index;default:0"`         // 状态：0-未开始 1-进行中 2-已结束 3-已取消
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SeckillActivityProduct 秒杀活动商品关联表（多对多）
type SeckillActivityProduct struct {
	ID         int64 `gorm:"primaryKey"`
	ActivityID int64 `gorm:"index;not null"` // 活动ID
	ProductID  int64 `gorm:"index;not null"` // 商品ID
	SeckillStock int64 `gorm:"not null"`     // 该商品在此活动中的秒杀库存
	CreatedAt  time.Time
}

// Repository 秒杀活动仓储接口
type Repository interface {
	// 活动CRUD
	Create(ctx context.Context, activity *SeckillActivity) error
	GetByID(ctx context.Context, id int64) (*SeckillActivity, error)
	ListAll(ctx context.Context) ([]*SeckillActivity, error)
	Update(ctx context.Context, activity *SeckillActivity) error
	Delete(ctx context.Context, id int64) error
	
	// 活动商品关联
	AddProduct(ctx context.Context, activityID, productID, seckillStock int64) error
	RemoveProduct(ctx context.Context, activityID, productID int64) error
	GetProductsByActivity(ctx context.Context, activityID int64) ([]*SeckillActivityProduct, error)
	GetActivitiesByProduct(ctx context.Context, productID int64) ([]*SeckillActivity, error)
}
