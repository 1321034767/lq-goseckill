package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/mq"
	"github.com/example/goseckill/internal/infra/redis"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
)

// RegisterAdminRoutes 注册后台管理端的 HTTP 路由
// 端口通常是 8081，与前台 Web 服务分离。
func RegisterAdminRoutes(app *iris.Application, cfg *config.Config) {
	// 初始化基础设施
	db := mysql.Init(&cfg.MySQL)
	redisClient := redis.Init(&cfg.Redis)
	mqConn := mq.Init(&cfg.RabbitMQ)

	// 仓储与服务
	productRepo := mysql.NewProductRepository(db)
	orderRepo := mysql.NewOrderRepository(db)
	userRepo := mysql.NewUserRepository(db)
	activityRepo := mysql.NewSeckillActivityRepository(db)
	chatRepo := mysql.NewChatRepository(db)

	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(orderRepo)
	chatSvc := service.NewChatService(chatRepo)
	accountSvc := service.NewAccountService(db, productRepo, orderRepo, userRepo)
	activitySvc := service.NewSeckillActivityService(activityRepo, productRepo)
	seckillSvc := service.NewSeckillService(productRepo, activityRepo, redisClient, mqConn, &cfg.JWT)

	// 静态资源
	app.HandleDir("/assets", iris.Dir("./web/admin/assets"))
	app.Get("/", func(ctx iris.Context) {
		_ = ctx.ServeFile("./web/admin/index.html")
	})

	api := app.Party("/api")

	// ---------- 商品管理 ----------

	// 商品列表（后台用：返回所有商品）
	api.Get("/products", func(ctx iris.Context) {
		list, err := productSvc.ListAll(ctx.Request().Context())
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// 创建商品
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

	// 更新商品
	api.Put("/products/{id:uint64}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		p, err := productSvc.GetByID(ctx.Request().Context(), int64(id))
		if err != nil {
			ctx.StopWithJSON(404, iris.Map{"code": 404, "msg": "product not found"})
			return
		}
		var req productRequest
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if err := req.applyTo(p, true); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		if err := productSvc.Update(ctx.Request().Context(), p); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": p})
	})

	// ---------- 订单管理 ----------

	// 最近订单列表
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

	// ---------- 用户余额 / 订单管理 ----------

	// 用户余额列表
	api.Get("/users", func(ctx iris.Context) {
		list, err := accountSvc.ListAccounts(ctx.Request().Context())
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// 给指定用户充值（单位：分）
	api.Post("/users/{id:uint64}/recharge", func(ctx iris.Context) {
		uid, _ := ctx.Params().GetUint64("id")
		var req struct {
			Amount int64 `json:"amount"`
		}
		if err := ctx.ReadJSON(&req); err != nil || req.Amount <= 0 {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "无效充值金额"})
			return
		}
		acc, err := accountSvc.Recharge(ctx.Request().Context(), int64(uid), req.Amount)
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": iris.Map{
			"user_id": uid,
			"balance": acc.Balance,
			"frozen":  acc.Frozen,
		}})
	})

	// 指定用户订单
	api.Get("/users/{id:uint64}/orders", func(ctx iris.Context) {
		uid, _ := ctx.Params().GetUint64("id")
		list, err := accountSvc.ListOrdersByUser(ctx.Request().Context(), int64(uid))
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// ---------- 秒杀活动管理 ----------

	// 获取所有活动列表
	api.Get("/seckill-activities", func(ctx iris.Context) {
		_ = activitySvc.CheckAndUpdateExpiredActivities(ctx.Request().Context())
		list, err := activitySvc.ListActivities(ctx.Request().Context())
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// 获取活动详情（包含商品列表）
	api.Get("/seckill-activities/{id:uint64}", func(ctx iris.Context) {
		_ = activitySvc.CheckAndUpdateExpiredActivities(ctx.Request().Context())
		id, _ := ctx.Params().GetUint64("id")
		data, err := activitySvc.GetActivity(ctx.Request().Context(), int64(id))
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": data})
	})

	// 创建秒杀活动
	api.Post("/seckill-activities", func(ctx iris.Context) {
		var req struct {
			Name          string          `json:"name"`
			Description   string          `json:"description"`
			StartTime     string          `json:"start_time"`
			EndTime       string          `json:"end_time"`
			Discount      float64         `json:"discount"`
			LimitPerUser  int64           `json:"limit_per_user"`
			ProductIDs    []int64         `json:"product_ids"`
			ProductStocks map[int64]int64 `json:"product_stocks"`
		}
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		start, err := parseAdminTime(req.StartTime)
		if err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "invalid start_time: " + err.Error()})
			return
		}
		end, err := parseAdminTime(req.EndTime)
		if err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "invalid end_time: " + err.Error()})
			return
		}
		if req.Discount <= 0 || req.Discount > 1 {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "discount must be between 0 and 1"})
			return
		}
		if req.LimitPerUser <= 0 {
			req.LimitPerUser = 1
		}
		activity, err := activitySvc.CreateActivity(ctx.Request().Context(), &service.CreateActivityRequest{
			Name:          req.Name,
			Description:   req.Description,
			StartTime:     start,
			EndTime:       end,
			Discount:      req.Discount,
			LimitPerUser:  req.LimitPerUser,
			ProductIDs:    req.ProductIDs,
			ProductStocks: req.ProductStocks,
		})
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": activity})
	})

	// 更新秒杀活动（不含商品列表）
	api.Put("/seckill-activities/{id:uint64}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		var req struct {
			Name         string  `json:"name"`
			Description  string  `json:"description"`
			StartTime    string  `json:"start_time"`
			EndTime      string  `json:"end_time"`
			Discount     float64 `json:"discount"`
			LimitPerUser int64   `json:"limit_per_user"`
		}
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		start, err := parseAdminTime(req.StartTime)
		if err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "invalid start_time: " + err.Error()})
			return
		}
		end, err := parseAdminTime(req.EndTime)
		if err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "invalid end_time: " + err.Error()})
			return
		}
		if req.Discount <= 0 || req.Discount > 1 {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": "discount must be between 0 and 1"})
			return
		}
		if err := activitySvc.UpdateActivity(ctx.Request().Context(), int64(id), &service.UpdateActivityRequest{
			Name:         req.Name,
			Description:  req.Description,
			StartTime:    start,
			EndTime:      end,
			Discount:     req.Discount,
			LimitPerUser: req.LimitPerUser,
		}); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": "ok"})
	})

	// 启动活动（更新商品状态并同步库存到 Redis）
	api.Post("/seckill-activities/{id:uint64}/start", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		if err := activitySvc.StartActivity(ctx.Request().Context(), int64(id), seckillSvc); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "msg": "activity started"})
	})

	// 删除秒杀活动
	api.Delete("/seckill-activities/{id:uint64}", func(ctx iris.Context) {
		id, _ := ctx.Params().GetUint64("id")
		if err := activitySvc.DeleteActivity(ctx.Request().Context(), int64(id)); err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "msg": "deleted"})
	})

	// ---------- 聊天示例接口 ----------

	api.Get("/chat/contacts", func(ctx iris.Context) {
		// 返回内置的联系人列表
		type contact struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			LastMessage string `json:"last_message"`
		}
		contacts := []contact{
			{ID: "claire", Name: "Claire Sassu", Status: "away", LastMessage: "Can you share the..."},
			{ID: "maggie", Name: "Maggie Jackson", Status: "online", LastMessage: "I confirmed the info."},
			{ID: "joel", Name: "Joel King", Status: "offline", LastMessage: "Ready for the meeting..."},
			{ID: "mike", Name: "Mike Bolthort", Status: "online"},
		}
		ctx.JSON(iris.Map{"code": 0, "data": contacts})
	})

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
		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

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

// ---- 辅助结构与函数 ----

type productRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	Stock        int64  `json:"stock"`
	SeckillStock int64  `json:"seckill_stock"`
	Category     string `json:"category"`
	Status       int    `json:"status"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

func (r *productRequest) applyTo(p *product.Product, keepEmptyTime bool) error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	p.Name = r.Name
	p.Description = r.Description
	p.Price = r.Price
	p.Stock = r.Stock
	p.SeckillStock = r.SeckillStock
	if r.Category != "" {
		p.Category = r.Category
	}
	p.Status = r.Status

	if r.StartTime != "" {
		if t, err := parseAdminTime(r.StartTime); err == nil {
			p.StartTime = t
		} else if !keepEmptyTime {
			return err
		}
	} else if !keepEmptyTime && p.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}

	if r.EndTime != "" {
		if t, err := parseAdminTime(r.EndTime); err == nil {
			p.EndTime = t
		} else if !keepEmptyTime {
			return err
		}
	} else if !keepEmptyTime && p.EndTime.IsZero() {
		return fmt.Errorf("end_time is required")
	}

	return nil
}

// 支持多种常见时间格式，精确到秒
func parseAdminTime(v string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", v)
}
