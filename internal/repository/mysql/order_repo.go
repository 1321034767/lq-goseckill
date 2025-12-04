package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/goseckill/internal/datamodels/order"
)

type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储
func NewOrderRepository(db *gorm.DB) order.Repository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, o *order.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *orderRepo) GetByID(ctx context.Context, id int64) (*order.Order, error) {
	var o order.Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) ListByUser(ctx context.Context, userID int64) ([]*order.Order, error) {
	var list []*order.Order
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderRepo) ListRecent(ctx context.Context, limit int) ([]*order.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []*order.Order
	if err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
