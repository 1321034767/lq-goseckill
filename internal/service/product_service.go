package service

import (
	"context"

	"github.com/example/goseckill/internal/datamodels/product"
)

type ProductService struct {
	repo product.Repository
}

func NewProductService(repo product.Repository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) ListOnline(ctx context.Context) ([]*product.Product, error) {
	return s.repo.ListOnline(ctx)
}

func (s *ProductService) ListAll(ctx context.Context) ([]*product.Product, error) {
	return s.repo.ListAll(ctx)
}

func (s *ProductService) ListByCategory(ctx context.Context, category string) ([]*product.Product, error) {
	return s.repo.ListByCategory(ctx, category)
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*product.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, p *product.Product) error {
	return s.repo.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *product.Product) error {
	return s.repo.Update(ctx, p)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
