package mysql

import (
	"log"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/chat"
	"github.com/example/goseckill/internal/datamodels/order"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/datamodels/user"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Init 初始化全局 GORM 实例并自动迁移表结构
func Init(cfg *config.MySQLConfig) *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect mysql: %v", err)
		}

		if err = db.AutoMigrate(&user.User{}, &product.Product{}, &order.Order{}, &chat.Message{}); err != nil {
			log.Fatalf("auto migrate failed: %v", err)
		}
	})
	return db
}

// DB 获取全局 DB
func DB() *gorm.DB {
	return db
}
