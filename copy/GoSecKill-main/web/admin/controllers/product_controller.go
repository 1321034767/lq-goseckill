package controllers

import (
	"GoSecKill/internal/services"
	"GoSecKill/pkg/models"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go.uber.org/zap"
)

type ProductController struct {
	productService services.IProductService
}

func NewProductController(productService services.IProductService) *ProductController {
	return &ProductController{productService: productService}
}

func (c *ProductController) GetProductList(ctx iris.Context) mvc.View {
	products, _ := c.productService.GetProductList()
	zap.L().Info("Get product list", zap.Any("products", products))
	return mvc.View{
		Name: "product/view.html",
		Data: iris.Map{
			"products": products,
		},
	}
}

func (c *ProductController) PostUpdate(ctx iris.Context) {
	product := models.Product{}

	if err := ctx.ReadForm(&product); err != nil {
		zap.L().Error("Failed to read form", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}

	err := c.productService.UpdateProduct(product)
	if err != nil {
		zap.L().Error("Failed to update product", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to update product")
		return
	}

	zap.L().Info("Successfully updated product", zap.Any("product", product))
	ctx.Redirect("/product/all")
}

func (c *ProductController) GetAddProduct(ctx iris.Context) mvc.View {
	return mvc.View{
		Name: "product/add.html",
	}
}

func (c *ProductController) PostProduct(ctx iris.Context) {
	product := models.Product{}

	if err := ctx.ReadForm(&product); err != nil {
		zap.L().Error("Failed to read form", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}

	err := c.productService.InsertProduct(product)
	if err != nil {
		zap.L().Error("Failed to insert product", zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to insert product")
		return
	}

	zap.L().Info("Successfully inserted product", zap.Any("product", product))
	ctx.Redirect("/product/all")
}

func (c *ProductController) GetManagerProduct(ctx iris.Context) mvc.View {
	idString := ctx.URLParam("id")
	if idString == "" {
		zap.L().Error("No product ID provided in request")
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "No product ID provided"})
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "No product ID provided",
			},
		}
	}

	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		zap.L().Error("Invalid product ID format", zap.String("id", idString), zap.Error(err))
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "Invalid ID format",
			},
		}
	}

	product, err := c.productService.GetProductByID(int(id))
	if err != nil {
		zap.L().Error("Failed to get product", zap.Int64("id", id), zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{"error": "Failed to get product"})
		return mvc.View{
			Name: "shared/error.html",
			Data: iris.Map{
				"Message": "Failed to get product",
			},
		}
	}

	return mvc.View{
		Name: "product/manager.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (c *ProductController) DeleteProduct(ctx iris.Context) {
	idString := ctx.URLParam("id")
	if idString == "" {
		zap.L().Error("No product ID provided in request")
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "No product ID provided"})
		return
	}

	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		zap.L().Error("Invalid product ID format", zap.String("id", idString), zap.Error(err))
		ctx.StatusCode(iris.StatusBadRequest)
		_, _ = ctx.JSON(iris.Map{"error": "Invalid ID format"})
		return
	}

	err = c.productService.DeleteProduct(int(id))
	if err != nil {
		zap.L().Error("Failed to delete product", zap.Int64("id", id), zap.Error(err))
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.JSON(iris.Map{"error": "Failed to delete product"})
		return
	}

	zap.L().Info("Successfully deleted product", zap.Int64("id", id))
	ctx.Redirect("/product/all")
}

func (c *ProductController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/all", "GetProductList")
	b.Handle("GET", "/manager", "GetManagerProduct")
	b.Handle("GET", "/add", "GetAddProduct")
	b.Handle("POST", "/add", "PostProduct")
	b.Handle("GET", "/delete", "DeleteProduct")
}
