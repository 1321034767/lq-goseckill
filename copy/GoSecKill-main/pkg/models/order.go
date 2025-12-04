package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserId     uint `gorm:"not null" json:"user_id"`
	ProductId  uint `gorm:"not null" json:"product_id"`
	ProductNum int  `gorm:"not null" json:"product_num"`
	Status     int  `gorm:"not null" json:"status"`
}

const (
	OrderStatusWait = iota
	OrderStatusSuccess
	OrderStatusFailed
)
