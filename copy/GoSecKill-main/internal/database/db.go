package database

import (
	"GoSecKill/pkg/models"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

// InitDB initializes the database connection
func InitDB() *gorm.DB {
	// Connect to the database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.database"),
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("failed to connect database", zap.Error(err))
	}

	// Auto migrate the database
	_ = db.AutoMigrate(&models.Product{}, &models.Order{}, &models.User{})

	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("failed to get database connection pool", zap.Error(err))
	}

	// Set the database connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

	return db
}

func GetDB() *gorm.DB {
	return db
}
