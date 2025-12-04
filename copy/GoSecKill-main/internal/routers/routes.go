package routers

import (
	"GoSecKill/internal/middleware"
	"GoSecKill/internal/services"
	"GoSecKill/pkg/mq"
	"GoSecKill/pkg/repositories"
	controllers2 "GoSecKill/web/admin/controllers"
	"GoSecKill/web/server/controllers"
	"context"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"gorm.io/gorm"
)

func InitAdminRoutes(app *iris.Application, db *gorm.DB, ctx context.Context) {
	productRepository := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepository)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	product.Register(ctx, productService)
	product.Handle(new(controllers2.ProductController))
	product.Handle(controllers2.NewProductController(productService))

	orderRepository := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(orderRepository)
	orderParty := app.Party("/order")
	order := mvc.New(orderParty)
	order.Register(ctx, orderService)
	order.Handle(new(controllers2.OrderController))
	order.Handle(controllers2.NewOrderController(orderService))
}

func InitServerRoutes(app *iris.Application, db *gorm.DB, ctx context.Context, sessions *sessions.Sessions, rabbitmq *mq.RabbitMQ) {
	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(ctx, userService)
	user.Handle(new(controllers.UserController))
	user.Handle(controllers.NewUserController(userService, sessions))

	product := repositories.NewProductRepository(db)
	productService := services.NewProductService(product)
	order := repositories.NewOrderRepository(db)
	orderService := services.NewOrderService(order)
	
	// DO NOT register /product route here - it will be registered in main.go after InitServerRoutes
	// to ensure it takes precedence over Party routes
	
	// Register product routes using Party (for /product/detail, /product/order, etc.)
	// Note: We don't register /product root path here to avoid conflicts
	proProduct := app.Party("/product")
	proProduct.Use(middleware.AuthConProduct)
	
	// Register specific routes manually
	proProduct.Get("/detail", func(ctx iris.Context) {
		productCtrl := controllers.NewProductController(productService, orderService, sessions)
		view := productCtrl.GetDetail(ctx)
		if view.Layout != "" {
			ctx.ViewLayout(view.Layout)
		}
		_ = ctx.View(view.Name, view.Data)
	})
	
	proProduct.Get("/order", func(ctx iris.Context) {
		productCtrl := controllers.NewProductController(productService, orderService, sessions)
		result := productCtrl.GetOrder(ctx)
		ctx.Write(result)
	})
	
	// Register MVC for other routes if needed
	pro := mvc.New(proProduct)
	pro.Register(productService, orderService, rabbitmq)
	productController := controllers.NewProductController(productService, orderService, sessions)
	pro.Handle(productController)
}
