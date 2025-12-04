package product

import (
	"context"
	"time"
)

// Product 商品模型
type Product struct {
	ID           int64     `gorm:"primaryKey"`
	Name         string    `gorm:"size:128;not null"`
	Description  string    `gorm:"size:512"`
	Price        int64     `gorm:"not null"` // 分
	Stock        int64     `gorm:"not null"`
	SeckillStock int64     `gorm:"not null"`
	Category     string    `gorm:"size:32;index"` // 分类：men(男士)、women(女士)、accessories(饰品)
	StartTime    time.Time `gorm:"index"`
	EndTime      time.Time `gorm:"index"`
	Status       int       `gorm:"index"` // 0:下线 1:正常 2:秒杀中
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Repository 商品仓储接口
type Repository interface {
	GetByID(ctx context.Context, id int64) (*Product, error)
	ListAll(ctx context.Context) ([]*Product, error)
	ListOnline(ctx context.Context) ([]*Product, error)
	ListByCategory(ctx context.Context, category string) ([]*Product, error) // 按分类查询
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
}


