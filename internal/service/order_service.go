package service

import (
	"context"

	"github.com/example/goseckill/internal/datamodels/order"
)

// OrderService 用于后台订单查询等场景
type OrderService struct {
	repo order.Repository
}

// NewOrderService 创建订单服务
func NewOrderService(repo order.Repository) *OrderService {
	return &OrderService{repo: repo}
}

// ListRecent 查询最新的订单记录
func (s *OrderService) ListRecent(ctx context.Context, limit int) ([]*order.Order, error) {
	return s.repo.ListRecent(ctx, limit)
}
