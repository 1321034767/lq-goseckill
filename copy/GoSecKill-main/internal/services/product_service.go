package services

import (
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/repositories"
)

type IProductService interface {
	GetProductList() (products []models.Product, err error)

	GetProductByID(id int) (product models.Product, err error)

	InsertProduct(product models.Product) (err error)

	UpdateProduct(product models.Product) (err error)

	DeleteProduct(id int) (err error)

	SubNumberOne(id int) (err error)
}

type ProductService struct {
	productRepository repositories.IProductRepository
}

func NewProductService(productRepository repositories.IProductRepository) IProductService {
	return &ProductService{productRepository: productRepository}
}

func (p ProductService) GetProductList() (products []models.Product, err error) {
	return p.productRepository.GetProductList()
}

func (p ProductService) GetProductByID(id int) (product models.Product, err error) {
	return p.productRepository.GetProductByID(id)
}

func (p ProductService) InsertProduct(product models.Product) (err error) {
	return p.productRepository.InsertProduct(product)
}

func (p ProductService) UpdateProduct(product models.Product) (err error) {
	return p.productRepository.UpdateProduct(product)
}

func (p ProductService) DeleteProduct(id int) (err error) {
	return p.productRepository.DeleteProduct(id)
}

func (p ProductService) SubNumberOne(id int) (err error) {
	return p.productRepository.SubNumberOne(id)
}
