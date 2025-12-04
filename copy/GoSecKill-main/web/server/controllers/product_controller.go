package controllers

import (
	"GoSecKill/internal/services"
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/mq"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"go.uber.org/zap"
)

type ProductController struct {
	productService services.IProductService
	orderService   services.IOrderService
	sessions       *sessions.Sessions
	rabbitMQ       *mq.RabbitMQ
}

func NewProductController(productService services.IProductService, orderService services.IOrderService, sessions *sessions.Sessions) *ProductController {
	return &ProductController{productService: productService, orderService: orderService, sessions: sessions}
}

var (
	htmlOutPath = "./fronted/web/htmlProductShow/"

	templatePath = "./fronted/web/views/template/"
)

func (p *ProductController) GetGenerateHtml(ctx iris.Context) {
	productString := ctx.URLParam("productID")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		zap.L().Error("Failed to convert productID", zap.Error(err))
	}

	contentTmp, err := template.ParseFiles(filepath.Join(templatePath, "product.html"))
	if err != nil {
		zap.L().Error("Failed to parse template", zap.Error(err))
	}

	fileName := filepath.Join(htmlOutPath, "htmlProduct.html")

	product, err := p.productService.GetProductByID(productID)
	if err != nil {
		zap.L().Error("Failed to get product", zap.Error(err))
	}

	generateStaticHtml(ctx, contentTmp, fileName, &product)
}

func generateStaticHtml(ctx iris.Context, template *template.Template, fileName string, product *models.Product) {
	if exist(fileName) {
		err := os.Remove(fileName)
		if err != nil {
			zap.L().Error("Failed to remove file", zap.Error(err))
		}
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		zap.L().Error("Failed to open file", zap.Error(err))
	}
	defer file.Close()
	template.Execute(file, &product)
}

func exist(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

// Get handles GET /product - shows product detail page
func (p *ProductController) Get(ctx iris.Context) mvc.View {
	product, err := p.productService.GetProductByID(1)
	if err != nil {
		zap.L().Error("Failed to get product", zap.Error(err))
		// If product not found, return error view
		return mvc.View{
			Layout: "shared/productLayout.html",
			Name:   "shared/error.html",
			Data: iris.Map{
				"message": "Product not found",
				"status":  404,
			},
		}
	}

	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/view.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (p *ProductController) GetDetail(ctx iris.Context) mvc.View {
	product, err := p.productService.GetProductByID(1)
	if err != nil {
		zap.L().Error("Failed to get product", zap.Error(err))
	}

	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/view.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (p *ProductController) GetOrder(ctx iris.Context) []byte {
	productString := ctx.URLParam("productID")
	userString := ctx.GetCookie("uid")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		zap.L().Error("Failed to convert productID", zap.Error(err))
	}
	userID, err := strconv.Atoi(userString)
	if err != nil {
		zap.L().Error("Failed to convert userID", zap.Error(err))
	}

	message := models.NewMessage(int64(productID), int64(userID))
	byteMessage, err := json.Marshal(message)
	if err != nil {
		zap.L().Error("Failed to marshal message", zap.Error(err))
	}

	p.rabbitMQ.PublishSimple(string(byteMessage))

	return []byte("true")
}

// BeforeActivation registers routes for the product controller
func (p *ProductController) BeforeActivation(b mvc.BeforeActivation) {
	// Explicitly register GET /product route (empty path means root of the party)
	b.Handle("GET", "/", "Get")
	b.Handle("GET", "/detail", "GetDetail")
	b.Handle("GET", "/order", "GetOrder")
	b.Handle("GET", "/generateHtml", "GetGenerateHtml")
}

