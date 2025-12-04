package main

import (
	"log"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/server"
)

func main() {
	cfg := config.DefaultConfig()

	app := iris.New()
	server.RegisterAdminRoutes(app, cfg)

	addr := cfg.AdminServer.Addr()
	log.Printf("admin server listening on %s", addr)
	if err := app.Run(iris.Addr(addr)); err != nil {
		log.Fatalf("failed to run admin server: %v", err)
	}
}
