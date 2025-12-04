package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/auth"
	"github.com/example/goseckill/internal/config"
	"github.com/example/goseckill/internal/datamodels/product"
	"github.com/example/goseckill/internal/infra/mq"
	"github.com/example/goseckill/internal/infra/redis"
	"github.com/example/goseckill/internal/repository/mysql"
	"github.com/example/goseckill/internal/service"
	webcontrollers "github.com/example/goseckill/web/controllers"
)

// RegisterRoutes 注册所有 HTTP 路由
func RegisterRoutes(app *iris.Application, cfg *config.Config) {
	// 初始化基础设施
	db := mysql.Init(&cfg.MySQL)
	redisClient := redis.Init(&cfg.Redis)
	mqConn := mq.Init(&cfg.RabbitMQ)

	// 静态资源：挂载前端静态文件（CSS/JS/图片）
	app.HandleDir("/assets", iris.Dir("./web/assets"))

	// 首页：使用 web/index.html 作为入口页
	app.Get("/", func(ctx iris.Context) {
		// 直接返回静态首页 HTML
		_ = ctx.ServeFile("./web/index.html")
	})

	// 仓储与服务
	userRepo := mysql.NewUserRepository(db)
	productRepo := mysql.NewProductRepository(db)
	orderRepo := mysql.NewOrderRepository(db)
	_ = orderRepo

	userSvc := service.NewUserService(userRepo, &cfg.JWT)
	productSvc := service.NewProductService(productRepo)
	seckillSvc := service.NewSeckillService(productRepo, redisClient, mqConn, &cfg.JWT)

	api := app.Party("/api")

	// 健康检查
	api.Get("/health", func(ctx iris.Context) {
		ctx.JSON(iris.Map{
			"code": 0,
			"msg":  "ok",
		})
	})

	// 用户注册/登录（简单示例）
	api.Post("/register", func(ctx iris.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		u, err := userSvc.Register(ctx.Request().Context(), req.Username, req.Password)
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": u})
	})

	api.Post("/login", func(ctx iris.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := ctx.ReadJSON(&req); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		token, err := userSvc.Login(ctx.Request().Context(), req.Username, req.Password)
		if err != nil {
			ctx.StopWithJSON(401, iris.Map{"code": 401, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": iris.Map{"token": token}})
	})

	// 需要登录的接口
	authAPI := api.Party("/", func(ctx iris.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.StopWithJSON(401, iris.Map{"code": 401, "msg": "missing token"})
			return
		}
		claims, err := auth.ParseToken(&cfg.JWT, token)
		if err != nil {
			ctx.StopWithJSON(401, iris.Map{"code": 401, "msg": "invalid token"})
			return
		}
		ctx.Values().Set("user_id", claims.UserID)
		ctx.Values().Set("username", claims.Username)
		ctx.Next()
	})

	// 商品列表（支持按分类筛选）
	authAPI.Get("/products", func(ctx iris.Context) {
		category := ctx.URLParam("category") // 获取分类参数：men, women, accessories, 或空（全部）
		keyword := ctx.URLParam("q")         // 搜索关键字（按名称模糊匹配）
		var list []*product.Product
		var err error
		if category != "" {
			list, err = productSvc.ListByCategory(ctx.Request().Context(), category)
		} else {
			list, err = productSvc.ListOnline(ctx.Request().Context())
		}
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}

		// 如果带有关键字，则在内存中按名称做简单过滤
		if keyword != "" {
			kw := strings.ToLower(keyword)
			filtered := make([]*product.Product, 0, len(list))
			for _, p := range list {
				if strings.Contains(strings.ToLower(p.Name), kw) {
					filtered = append(filtered, p)
				}
			}
			list = filtered
		}

		ctx.JSON(iris.Map{"code": 0, "data": list})
	})

	// 获取秒杀路径
	authAPI.Get("/seckill/{id:uint64}/path", func(ctx iris.Context) {
		pid, _ := ctx.Params().GetUint64("id")
		userID := ctx.Values().GetInt64Default("user_id", 0)
		path, err := seckillSvc.GeneratePath(ctx.Request().Context(), userID, int64(pid))
		if err != nil {
			ctx.StopWithJSON(500, iris.Map{"code": 500, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "data": iris.Map{"path": path}})
	})

	// 发起秒杀
	authAPI.Post("/seckill/{id:uint64}/{path:string}", func(ctx iris.Context) {
		pid, _ := ctx.Params().GetUint64("id")
		path := ctx.Params().Get("path")
		userID := ctx.Values().GetInt64Default("user_id", 0)
		if err := seckillSvc.Seckill(ctx.Request().Context(), userID, int64(pid), path); err != nil {
			ctx.StopWithJSON(400, iris.Map{"code": 400, "msg": err.Error()})
			return
		}
		ctx.JSON(iris.Map{"code": 0, "msg": "queued"})
	})

	// 简单订单查询占位（实际逻辑可后续补充）
	authAPI.Get("/orders", func(ctx iris.Context) {
		userID := ctx.Values().GetInt64Default("user_id", 0)
		ctx.JSON(iris.Map{
			"code": 0,
			"msg":  "order list placeholder for user " + strconv.FormatInt(userID, 10),
		})
	})

	// ---------------- 前台页面路由 ----------------

	// 根据商品ID生成5种图片的映射（用于详情页大图）
	// 每个商品使用自己的专属图片（product_X.jpg 和 product_back_X.jpg）
	getProductImages := func(productID int64) map[string][]string {
		// 商品ID 1-12 直接对应图片 1-12
		// 商品ID 1 -> 图片1, ID 2 -> 图片2, ..., ID 12 -> 图片12
		var imgIndex int
		if productID >= 1 && productID <= 12 {
			imgIndex = int(productID) // ID 1->1, 2->2, ..., 12->12
		} else {
			// 对于其他ID，使用循环映射（1-12循环）
			imgIndex = int(((productID - 1) % 12) + 1)
		}
		
		// 图片现在存储在 /assets/img/shop/{imgIndex}/ 文件夹中
		folderPath := "/assets/img/shop/" + strconv.Itoa(imgIndex) + "/"
		
		var lg []string
		var thumb []string
		
		// 特殊处理：商品ID 6使用item_lg_1到item_lg_5图片
		if productID == 6 {
			lg = []string{
				folderPath + "item_lg_1.jpg",
				folderPath + "item_lg_2.jpg",
				folderPath + "item_lg_3.jpg",
				folderPath + "item_lg_4.jpg",
				folderPath + "item_lg_5.jpg",
			}
			thumb = []string{
				folderPath + "item_lg_1.jpg",
				folderPath + "item_lg_2.jpg",
				folderPath + "item_lg_3.jpg",
				folderPath + "item_lg_4.jpg",
				folderPath + "item_lg_5.jpg",
			}
		} else {
			// 其他商品使用product_X.jpg和product_back_X.jpg
			// 为每个商品生成5种图片：
			// 1. 主图（product_X.jpg）
			// 2. 背面图（product_back_X.jpg）
			// 3. 主图重复（用于展示不同角度/细节）
			// 4. 背面图重复（用于展示不同角度/细节）
			// 5. 主图再次重复（用于展示整体效果）
			// 注意：由于我们只有每个商品的2张图片，这里使用重复来创建5种视图
			// 实际应用中，应该为每个商品准备5张不同的图片
			basePath := folderPath + "product_"
			backPath := folderPath + "product_back_"
			
			lg = []string{
				basePath + strconv.Itoa(imgIndex) + ".jpg",
				backPath + strconv.Itoa(imgIndex) + ".jpg",
				basePath + strconv.Itoa(imgIndex) + ".jpg",
				backPath + strconv.Itoa(imgIndex) + ".jpg",
				basePath + strconv.Itoa(imgIndex) + ".jpg",
			}
			
			thumb = []string{
				basePath + strconv.Itoa(imgIndex) + ".jpg",
				backPath + strconv.Itoa(imgIndex) + ".jpg",
				basePath + strconv.Itoa(imgIndex) + ".jpg",
				backPath + strconv.Itoa(imgIndex) + ".jpg",
				basePath + strconv.Itoa(imgIndex) + ".jpg",
			}
		}
		
		return map[string][]string{
			"lg":    lg,
			"thumb": thumb,
		}
	}

	// 生成商品卡片所需的主图 / 背面图路径（用于列表 / 关联商品）
	type RelatedCard struct {
		Product *product.Product
		MainImg string
		BackImg string
	}

	getCardImages := func(p *product.Product) RelatedCard {
		// 复用与上面相同的图片索引规则
		var imgIndex int
		if p.ID >= 1 && p.ID <= 12 {
			imgIndex = int(p.ID)
		} else {
			imgIndex = int(((p.ID - 1) % 12) + 1)
		}

		folderPath := "/assets/img/shop/" + strconv.Itoa(imgIndex) + "/"

		// 默认使用 product_X / product_back_X
		mainImg := folderPath + "product_" + strconv.Itoa(imgIndex) + ".jpg"
		backImg := folderPath + "product_back_" + strconv.Itoa(imgIndex) + ".jpg"

		// 特殊处理：商品ID 6 使用 item_lg_1 / item_lg_2
		if p.ID == 6 {
			mainImg = folderPath + "item_lg_1.jpg"
			backImg = folderPath + "item_lg_2.jpg"
		}

		return RelatedCard{
			Product: p,
			MainImg: mainImg,
			BackImg: backImg,
		}
	}

	// 商品详情页：/product/{id}
	app.Get("/product/{id:uint64}", func(ctx iris.Context) {
		pid, _ := ctx.Params().GetUint64("id")
		p, err := productSvc.GetByID(ctx.Request().Context(), int64(pid))
		if err != nil {
			// 记录错误日志
			ctx.Application().Logger().Warnf("查询商品失败 (ID: %d): %v", pid, err)
			// 即使商品不存在，也使用 productLayout 以保持一致的页面结构
			ctx.ViewLayout("shared/productLayout.html")
			_ = ctx.View("shared/error.html", iris.Map{
				"showMessage": fmt.Sprintf("商品不存在或已下线 (ID: %d)", pid),
				"orderID":     "",
			})
			return
		}
		if p == nil {
			ctx.Application().Logger().Warnf("商品不存在 (ID: %d)", pid)
			ctx.ViewLayout("shared/productLayout.html")
			_ = ctx.View("shared/error.html", iris.Map{
				"showMessage": fmt.Sprintf("商品不存在或已下线 (ID: %d)", pid),
				"orderID":     "",
			})
			return
		}

		images := getProductImages(p.ID)

		// 面包屑中使用的商品分类中文名称
		categoryLabel := "商品详情"
		switch p.Category {
		case "men":
			categoryLabel = "男士"
		case "women":
			categoryLabel = "女士"
		case "accessories":
			categoryLabel = "饰品"
		}

		// 计算 “Shop the look” 关联商品：当前商品之后的 11 个商品（循环，除去自己共 11 个）
		var relatedCards []RelatedCard
		const maxRelated = 11
		for i := int64(1); i <= maxRelated; i++ {
			// 下一个商品ID（1-12 循环）
			nextID := ((p.ID-1+i)%12 + 1)
			if nextID == p.ID {
				continue
			}
			rp, err := productSvc.GetByID(ctx.Request().Context(), nextID)
			if err != nil || rp == nil {
				continue
			}
			// 只展示状态为正常的商品
			if rp.Status != 1 {
				continue
			}
			relatedCards = append(relatedCards, getCardImages(rp))
			if len(relatedCards) >= maxRelated {
				break
			}
		}

		ctx.ViewLayout("shared/productLayout.html")
		if err := ctx.View("product/view.html", iris.Map{
			"product":        p,
			"images":         images,
			"related_cards":  relatedCards,
			"category_label": categoryLabel,
		}); err != nil {
			ctx.Application().Logger().Errorf("渲染商品详情页失败: %v", err)
			ctx.ViewLayout("shared/productLayout.html")
			_ = ctx.View("shared/error.html", iris.Map{
				"showMessage": "页面渲染失败: " + err.Error(),
				"orderID":     "",
			})
			return
		}
	})

	// /product 默认跳转到一个示例商品（ID=1），实际可以改成商品列表页
	app.Get("/product", func(ctx iris.Context) {
		ctx.Redirect("/product/1", iris.StatusFound)
	})

	// 用户登录 / 注册表单路由
	userController := webcontrollers.NewUserController(userSvc)
	app.Get("/login", userController.ShowLogin)
	app.Get("/register", userController.ShowRegister)
	app.Get("/user/login", userController.ShowLogin)
	app.Get("/user/register", userController.ShowRegister)
	app.Get("/user/manage", userController.ShowManage)
	app.Get("/user/logout", userController.Logout)
	app.Post("/user/login", userController.PostLogin)
	app.Post("/user/add", userController.PostAdd)
}
