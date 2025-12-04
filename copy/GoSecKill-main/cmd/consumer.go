package main

import (
	"GoSecKill/internal/config"
	"GoSecKill/internal/database"
	"GoSecKill/internal/services"
	"GoSecKill/pkg/log"
	"GoSecKill/pkg/mq"
	"GoSecKill/pkg/repositories"

	"go.uber.org/zap"
)

func main() {
	// Load application configuration
	if err := config.LoadConfig("./config"); err != nil {
		panic(err)
	}

	// Initialize logger
	log.InitLogger()
	zap.L().Info("log init success")

	// Initialize database
	db := database.InitDB()

	productRepository := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepository)
	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository)

	rabbitmqConsumer := mq.NewRabbitMQSimple("go_seckill")
	rabbitmqConsumer.ConsumeSimple(orderService, productService)
}
