package main

import (
	"fmt"
	"log"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/server"
)

func main() {
	// 加载配置（目前使用默认配置，可后续扩展为从文件读取）
	cfg := config.DefaultConfig()

	app := iris.New()
	// 注册 HTML 模板引擎，使用本项目下的 web/views 目录
	// 注意：不直接依赖 copy/GoSecKill-main，而是只复制其中的前端模板到本项目
	tmpl := iris.HTML("./web/views", ".html")
	tmpl.Reload(true) // 开发模式下启用热重载，方便调试
	
	// 添加价格格式化函数：将分转换为美元格式（例如 990 -> $9.90）
	tmpl.AddFunc("formatPrice", func(price int64) string {
		return fmt.Sprintf("$%.2f", float64(price)/100.0)
	})
	
	app.RegisterView(tmpl)

	server.RegisterRoutes(app, cfg)

	addr := cfg.Server.Addr()
	log.Printf("web server listening on %s", addr)
	if err := app.Run(iris.Addr(addr)); err != nil {
		log.Fatalf("failed to run web server: %v", err)
	}
}


