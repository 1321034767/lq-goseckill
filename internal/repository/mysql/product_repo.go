package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/example/goseckill/internal/datamodels/product"
)

type productRepo struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储
func NewProductRepository(db *gorm.DB) product.Repository {
	return &productRepo{db: db}
}

func (r *productRepo) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	var p product.Product
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepo) ListAll(ctx context.Context) ([]*product.Product, error) {
	var list []*product.Product
	if err := r.db.WithContext(ctx).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *productRepo) ListOnline(ctx context.Context) ([]*product.Product, error) {
	var list []*product.Product
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *productRepo) ListByCategory(ctx context.Context, category string) ([]*product.Product, error) {
	var list []*product.Product
	query := r.db.WithContext(ctx).Where("status = ?", 1)
	if category != "" && category != "all" {
		// 查询指定分类的商品，category字段必须匹配
		query = query.Where("category = ?", category)
	}
	if err := query.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *productRepo) Create(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productRepo) Update(ctx context.Context, p *product.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&product.Product{}, id).Error
}
