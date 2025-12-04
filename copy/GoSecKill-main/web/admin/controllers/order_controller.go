package controllers

import (
	"GoSecKill/internal/services"
	"GoSecKill/pkg/models"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.uber.org/zap"
)

type OrderController struct {
	orderService services.IOrderService
}

func NewOrderController(orderService services.IOrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

func (c *OrderController) GetOrderList() mvc.View {
	orders, _ := c.orderService.GetOrderList()
	return mvc.View{
		Name: "order/view.html",
		Data: iris.Map{
			"orders": orders,
		},
	}
}

func (c *OrderController) PostUpdate(ctx iris.Context) {
	order := models.Order{}

	if err := ctx.ReadForm(&order); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}

	err := c.orderService.UpdateOrder(order)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to update order")
		return
	}

	ctx.Redirect("/order/all")
}

func (c *OrderController) GetAddOrder(ctx iris.Context) mvc.View {
	return mvc.View{
		Name: "order/add.html",
	}
}

func (c *OrderController) PostOrder(ctx iris.Context) {
	order := models.Order{}

	if err := ctx.ReadForm(&order); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}

	_, err := c.orderService.InsertOrder(order)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to insert order")
		return
	}

	ctx.Redirect("/order/all")
}

func (c *OrderController) GetManagerOrder(ctx iris.Context) mvc.View {
	idString := ctx.URLParam("id")
	if idString == "" {
		zap.L().Error("No order ID provided in request")
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "No order ID provided"})
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "No order ID provided",
			},
		}
	}

	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		zap.L().Error("Failed to parse order ID", zap.Error(err))
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "Invalid order ID"})
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "Invalid order ID",
			},
		}
	}

	order, err := c.orderService.GetOrderByID(int(id))
	if err != nil {
		zap.L().Error("Failed to get order by ID", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to get order by ID")
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "Failed to get order by ID",
			},
		}
	}

	return mvc.View{
		Name: "order/manager.html",
		Data: iris.Map{
			"order": order,
		},
	}
}

func (c *OrderController) DeleteOrder(ctx iris.Context) {
	idString := ctx.URLParam("id")
	if idString == "" {
		zap.L().Error("No order ID provided in request")
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "No order ID provided"})
		return
	}

	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		zap.L().Error("Failed to parse order ID", zap.Error(err))
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "Invalid order ID"})
		return
	}

	err = c.orderService.DeleteOrder(int(id))
	if err != nil {
		zap.L().Error("Failed to delete order", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to delete order")
		return
	}

	ctx.Redirect("/order/all")
}

func (c *OrderController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/all", "GetOrderList")
	b.Handle("POST", "/update", "PostUpdate")
	b.Handle("GET", "/add", "GetAddOrder")
	b.Handle("POST", "/add", "PostOrder")
	//b.Handle("GET", "/manager", "GetManagerOrder")
	b.Handle("GET", "/delete", "DeleteOrder")
}
