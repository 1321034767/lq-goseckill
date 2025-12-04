package main

import (
	"GoSecKill/internal/config"
	"GoSecKill/internal/database"
	"GoSecKill/internal/middleware"
	"GoSecKill/internal/routers"
	"GoSecKill/internal/services"
	"GoSecKill/pkg/log"
	"GoSecKill/pkg/mq"
	"GoSecKill/pkg/repositories"
	"GoSecKill/web/server/controllers"
	"context"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/spf13/viper"
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

	// Initialize the web application
	app := iris.New()
	app.Logger().SetLevel("debug")

	// Register the view engine
	template := iris.HTML("./web/server/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)

	// Add root path route first - redirect to login page
	app.Get("/", func(ctx iris.Context) {
		ctx.Redirect("/user/login")
	})

	// Register the routes
	app.HandleDir("/html", "./web/server/htmlProductShow")
	app.HandleDir("/assets", "./web/server/assets")
	
	app.OnAnyErrorCode(func(ctx iris.Context) {
		// Handle root path 404 - redirect to login
		if ctx.GetStatusCode() == 404 && ctx.Path() == "/" {
			ctx.Redirect("/user/login")
			return
		}
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "There was something wrong with the request!"))
		ctx.ViewData("status", ctx.GetStatusCode())
		ctx.ViewLayout("")
		_ = ctx.View("shared/error.html")
	})

	// Initialize the message queue
	rabbitmq := mq.NewRabbitMQSimple("go_seckill")

	session := sessions.New(sessions.Config{
		Cookie:  "sessioncookie",
		Expires: 24 * 60 * 60,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register the routes
	routers.InitServerRoutes(app, db, ctx, session, rabbitmq)
	
	// Register /product route AFTER InitServerRoutes to ensure it's registered last
	// This will override any Party routes that might conflict
	app.Get("/product", middleware.AuthConProduct, func(ctx iris.Context) {
		// Get services
		productRepo := repositories.NewProductRepository(db)
		productService := services.NewProductService(productRepo)
		orderRepo := repositories.NewOrderRepository(db)
		orderService := services.NewOrderService(orderRepo)
		
		productCtrl := controllers.NewProductController(productService, orderService, session)
		view := productCtrl.Get(ctx)
		
		if view.Layout != "" {
			ctx.ViewLayout(view.Layout)
		}
		
		if err := ctx.View(view.Name, view.Data); err != nil {
			zap.L().Error("Failed to render product view", zap.Error(err))
			ctx.StatusCode(500)
			ctx.ViewData("message", "Failed to load product page: "+err.Error())
			ctx.ViewData("status", 500)
			ctx.ViewLayout("")
			_ = ctx.View("shared/error.html")
		}
	})
	
	// Debug: Print all registered routes for /product
	zap.L().Info("=== Registered Routes Debug ===")
	routes := app.GetRoutes()
	productRoutesFound := false
	for _, r := range routes {
		if strings.Contains(r.Path, "product") {
			productRoutesFound = true
			zap.L().Info("Product route", 
				zap.String("method", r.Method), 
				zap.String("path", r.Path), 
				zap.String("name", r.Name))
		}
	}
	if !productRoutesFound {
		zap.L().Warn("No /product routes found! Listing all routes:")
		for _, r := range routes {
			zap.L().Info("Route", zap.String("method", r.Method), zap.String("path", r.Path))
		}
	} else {
		zap.L().Info("Product routes registered successfully")
	}
	zap.L().Info("=== End Routes Debug ===")

	// Start the web application
	err := app.Run(
		iris.Addr(viper.GetString("server.port")),
		iris.WithCharset("UTF-8"),
		iris.WithOptimizations,
		iris.WithoutServerError(iris.ErrServerClosed),
	)
	if err != nil {
		zap.L().Fatal("app run failed", zap.Error(err))
	}
}
