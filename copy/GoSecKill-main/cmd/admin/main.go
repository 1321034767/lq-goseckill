package main

import (
	"GoSecKill/internal/config"
	"GoSecKill/internal/database"
	"GoSecKill/internal/routers"
	"GoSecKill/pkg/log"
	"context"

	"github.com/kataras/iris/v12"
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
	template := iris.HTML("./web/admin/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)

	app.Get("/", func(ctx iris.Context) {
		_ = ctx.View("static/index.html")
	})

	// Register the routes
	app.HandleDir("/assets", "./web/admin/assets")
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "There was something wrong with the request!"))
		ctx.ViewData("status", ctx.GetStatusCode())
		ctx.ViewLayout("")
		_ = ctx.View("shared/error.html")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register the routes
	routers.InitAdminRoutes(app, db, ctx)

	// Start the web application
	err := app.Run(
		iris.Addr(viper.GetString("server.adminPort")),
		iris.WithCharset("UTF-8"),
		iris.WithOptimizations,
		iris.WithoutServerError(iris.ErrServerClosed),
	)
	if err != nil {
		zap.L().Fatal("app run failed", zap.Error(err))
	}
}
