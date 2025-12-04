package repositories

import (
	"GoSecKill/pkg/models"

	"gorm.io/gorm"
)

type IProductRepository interface {
	GetProductList() (products []models.Product, err error)

	GetProductByID(id int) (product models.Product, err error)

	InsertProduct(product models.Product) (err error)

	UpdateProduct(product models.Product) (err error)

	DeleteProduct(id int) (err error)

	SubNumberOne(id int) (err error)
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) IProductRepository {
	return &ProductRepository{db: db}
}

func (p ProductRepository) GetProductList() (products []models.Product, err error) {
	var productList []models.Product
	err = p.db.Find(&productList).Error
	if err != nil {
		return nil, err
	}
	return productList, nil
}

func (p ProductRepository) GetProductByID(id int) (product models.Product, err error) {
	var productItem models.Product
	err = p.db.First(&productItem, id).Error
	if err != nil {
		return models.Product{}, err
	}
	return productItem, nil
}

func (p ProductRepository) InsertProduct(product models.Product) (err error) {
	err = p.db.Create(&product).Error
	if err != nil {
		return err
	}
	return nil
}

func (p ProductRepository) UpdateProduct(product models.Product) (err error) {
	err = p.db.Save(&product).Error
	if err != nil {
		return err
	}
	return nil
}

func (p ProductRepository) DeleteProduct(id int) (err error) {
	err = p.db.Delete(&models.Product{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (p ProductRepository) SubNumberOne(id int) (err error) {
	err = p.db.Model(&models.Product{}).Where("id = ? and number > 0", id).Update("number", gorm.Expr("number - ?", 1)).Error
	if err != nil {
		return err
	}
	return nil
}
