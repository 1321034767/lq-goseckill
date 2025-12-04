package services

import (
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/repositories"
)

type IOrderService interface {
	GetOrderList() (orders []models.Order, err error)

	GetOrderByID(id int) (order models.Order, err error)

	InsertOrder(order models.Order) (*models.Order, error)

	UpdateOrder(order models.Order) (err error)

	DeleteOrder(id int) (err error)

	InsertOrderByMessage(message *models.Message) (uint, error)
}

type OrderService struct {
	orderRepository repositories.IOrderRepository
}

func NewOrderService(orderRepository repositories.IOrderRepository) IOrderService {
	return &OrderService{orderRepository: orderRepository}
}

func (s OrderService) GetOrderList() (orders []models.Order, err error) {
	return s.orderRepository.GetOrderList()
}

func (s OrderService) GetOrderByID(id int) (order models.Order, err error) {
	return s.orderRepository.GetOrderByID(id)
}

func (s OrderService) InsertOrder(order models.Order) (*models.Order, error) {
	return s.orderRepository.InsertOrder(order)
}

func (s OrderService) UpdateOrder(order models.Order) (err error) {
	return s.orderRepository.UpdateOrder(order)
}

func (s OrderService) DeleteOrder(id int) (err error) {
	return s.orderRepository.DeleteOrder(id)
}

func (s OrderService) InsertOrderByMessage(message *models.Message) (orderId uint, err error) {
	order := &models.Order{
		UserId:     uint(message.UserID),
		ProductId:  uint(message.ProductID),
		ProductNum: 1,
		Status:     models.OrderStatusSuccess,
	}
	order, err = s.InsertOrder(*order)
	return order.ID, err
}
