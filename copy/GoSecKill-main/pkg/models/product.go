package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name    string  `gorm:"type:varchar(100);not null" json:"name"`
	Price   float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock   int     `gorm:"type:int;not null" json:"stock"`
	ImgPath string  `gorm:"type:varchar(255);not null" json:"img_path"`
	Detail  string  `gorm:"type:varchar(255);not null" json:"detail"`
}
