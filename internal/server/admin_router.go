package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

// RegisterAdminRoutes 注册后台管理相关的 HTTP 路由
func RegisterAdminRoutes(app *iris.Application, cfg *config.Config) {
	db := mysql.Init(&cfg.MySQL)

	productRepo := mysql.NewProductRepository(db)
	orderRepo := mysql.NewOrderRepository(db)
	chatRepo := mysql.NewChatRepository(db)

	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(orderRepo)
	chatSvc := service.NewChatService(chatRepo)

	// 静态资源与入口页
	app.HandleDir("/assets", iris.Dir("./web/admin/assets"))
	app.Get("/", func(ctx iris.Context) {
		_ = ctx.ServeFile("./web/admin/index.html")
	})

	api := app.Party("/api")

	api.Get("/health", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"code": 0, "msg": "ok"})
	})

	api.Get("/products", func(ctx iris.Context) {
		list, err := productSvc.ListAll(ctx.Request().Context())
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	api.Post("/products", func(ctx iris.Context) {
		var req productRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		p := &product.Product{}
		if err := req.applyTo(p, false); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if err := productSvc.Create(ctx.Request().Context(), p); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": p})
	})

	api.Put("/products/{id:uint64}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		existing, err := productSvc.GetByID(ctx.Request().Context(), int64(id))
		if err != nil {
			ctx.StopWithJSON(404, iris.Map{"code": 404, "msg": fmt.Sprintf("product %d not found", id)})
			return
		}
		var req productRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if err := req.applyTo(existing, true); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if err := productSvc.Update(ctx.Request().Context(), existing); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": existing})
	})

	api.Delete("/products/{id:uint64}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		if err := productSvc.Delete(ctx.Request().Context(), int64(id)); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "msg": "deleted"})
	})

	api.Get("/orders", func(ctx iris.Context) {
		limitStr := ctx.URLParamDefault("limit", "20")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 20
		}
		list, err := orderSvc.ListRecent(ctx.Request().Context(), limit)
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// -------------------- 聊天相关接口 --------------------

	// /api/chat/contacts 返回固定的联系人列表（前端展示用）
	api.Get("/chat/contacts", func(ctx iris.Context) {
		type contact struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			LastMessage string `json:"last_message"`
		}
		// 为了简单起见，联系人列表先固定写死，重点演示消息存储与收发
		contacts := []contact{
			{ID: "claire", Name: "Claire Sassu", Status: "away", LastMessage: "Can you share the..."},
			{ID: "maggie", Name: "Maggie Jackson", Status: "online", LastMessage: "I confirmed the info."},
			{ID: "joel", Name: "Joel King", Status: "offline", LastMessage: "Ready for the meeting..."},
			{ID: "mike", Name: "Mike Bolthort", Status: "online"},
			{ID: "maggie2", Name: "Maggie Jackson", Status: "online"},
			{ID: "jhon", Name: "Jhon Voltemar", Status: "offline"},
		}
		ctx.JSON(iris.Map{"code": 0, "data": contacts})
	})

	// /api/chat/messages/{id} 获取某个会话的消息列表，可通过 after_id 增量拉取
	api.Get("/chat/messages/{id:string}", func(ctx iris.Context) {
		contactID := ctx.Params().GetString("id")
		afterIDStr := ctx.URLParamDefault("after_id", "0")
		limitStr := ctx.URLParamDefault("limit", "50")

		var afterID uint64
		if v, err := strconv.ParseUint(afterIDStr, 10, 64); err == nil {
			afterID = v
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 50
		}

		list, err := chatSvc.ListMessages(ctx.Request().Context(), contactID, afterID, limit)
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}

		// 如果还没有任何消息，为特定联系人生成一段示例对话，方便前端展示气泡
		if len(list) == 0 && afterID == 0 {
			switch contactID {
			case "claire":
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "friend", "Hello")
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "self", "Hi, how are you?")
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "friend", "Good, I'll need support with my pc")
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "self", "Sure, just tell me what is going on with your computer?")
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "friend", "I don't know it just turns off suddenly")
			case "maggie":
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "friend", "I confirmed the info.")
			case "joel":
				_, _ = chatSvc.SendMessage(ctx.Request().Context(), contactID, "friend", "Ready for the meeting...")
			}
			// 重新查询一次，返回刚生成的示例消息
			list, _ = chatSvc.ListMessages(ctx.Request().Context(), contactID, 0, limit)
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// /api/chat/messages/{id} 发送一条消息（from 固定为 "self"，代表当前 Admin）
	api.Post("/chat/messages/{id:string}", func(ctx iris.Context) {
		contactID := ctx.Params().GetString("id")
		var req struct {
			Content string `json:"content"`
		}
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if req.Content == "" {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "content is empty"})
			return
		}
		m, err := chatSvc.SendMessage(ctx.Request().Context(), contactID, "self", req.Content)
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": m})
	})
}

type productRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Category     string `json:"category"` // 分类：men(男士)、women(女士)、accessories(饰品)
	Status       int    `json:"status"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

func (r *productRequest) applyTo(p *product.Product, keepEmptyTime bool) error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if r.Stock < 0 || r.SeckillStock < 0 {
		return fmt.Errorf("stock must be >= 0")
	}
	if r.SeckillStock > r.Stock {
		return fmt.Errorf("seckill_stock cannot exceed stock")
	}
	if r.Status < 0 || r.Status > 2 {
		return fmt.Errorf("status must be 0/1/2")
	}

	p.Name = r.Name
	p.Description = r.Description
	p.Price = r.Price
	p.Stock = r.Stock
	p.SeckillStock = r.SeckillStock
	p.Category = r.Category // 设置分类
	p.Status = r.Status

	if r.StartTime != "" {
		t, err := parseAdminTime(r.StartTime)
		if err != nil {
			return err
		}
		p.StartTime = t
	} else if !keepEmptyTime && p.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}

	if r.EndTime != "" {
		t, err := parseAdminTime(r.EndTime)
		if err != nil {
			return err
		}
		p.EndTime = t
	} else if !keepEmptyTime && p.EndTime.IsZero() {
		return fmt.Errorf("end_time is required")
	}

	return nil
}

func parseAdminTime(v string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", v)
}
