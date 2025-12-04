package order

import (
	"context"
	"time"
)

// Order 订单模型
type Order struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"index;not null"`
	ProductID int64     `gorm:"index;not null"`
	Price     int64     `gorm:"not null"`
	Status    int       `gorm:"index;not null"` // 0:已创建 1:已支付 2:已取消
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository 订单仓储接口
type Repository interface {
	Create(ctx context.Context, o *Order) error
	GetByID(ctx context.Context, id int64) (*Order, error)
	ListByUser(ctx context.Context, userID int64) ([]*Order, error)
	ListRecent(ctx context.Context, limit int) ([]*Order, error)
}


