package repositories

import (
	"GoSecKill/pkg/models"

	"gorm.io/gorm"
)

type IOrderRepository interface {
	GetOrderList() (orders []models.Order, err error)

	GetOrderByID(id int) (order models.Order, err error)

	InsertOrder(order models.Order) (*models.Order, error)

	UpdateOrder(order models.Order) (err error)

	DeleteOrder(id int) (err error)
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) IOrderRepository {
	return &OrderRepository{db: db}
}

func (o OrderRepository) GetOrderList() (orders []models.Order, err error) {
	var orderList []models.Order
	err = o.db.Find(&orderList).Error
	if err != nil {
		return nil, err
	}
	return orderList, nil
}

func (o OrderRepository) GetOrderByID(id int) (order models.Order, err error) {
	var orderItem models.Order
	err = o.db.First(&orderItem, id).Error
	if err != nil {
		return models.Order{}, err
	}
	return orderItem, nil
}

func (o OrderRepository) InsertOrder(order models.Order) (*models.Order, error) {
	err := o.db.Create(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (o OrderRepository) UpdateOrder(order models.Order) (err error) {
	err = o.db.Save(&order).Error
	if err != nil {
		return err
	}
	return nil
}

func (o OrderRepository) DeleteOrder(id int) (err error) {
	err = o.db.Delete(&models.Order{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
