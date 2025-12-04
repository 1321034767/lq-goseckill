package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"

	"github.com/example/goseckill/internal/service"
)

// ProductController 前台商品相关页面控制器（MVC）
// 路由在 internal/server/router.go 中通过 Iris MVC 挂载。
type ProductController struct {
	Ctx            iris.Context
	ProductService *service.ProductService
}

// GetBy 处理 GET /product/{id:uint64}
// 显示商品详情页，使用 productLayout 布局和 product/view.html 模板。
func (c *ProductController) GetBy(id uint64) mvc.View {
	p, err := c.ProductService.GetByID(c.Ctx.Request().Context(), int64(id))
	if err != nil || p == nil {
		return mvc.View{
			Layout: "shared/layout.html",
			Name:   "shared/error.html",
			Data: iris.Map{
				"showMessage": "商品不存在或已下线",
				"orderID":     "",
			},
		}
	}

	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/view.html",
		Data: iris.Map{
			"product": p,
		},
	}
}

// Get 映射到 GET /product
// 这里简单跳转到一个默认的商品 ID，实际项目中可做商品列表页。
func (c *ProductController) Get() {
	// 默认跳到 ID=1 的商品详情，避免 404。
	c.Ctx.Redirect("/product/1", iris.StatusFound)
}
