# 路由配置问题解决方案

## 问题描述

登录功能正常工作，但访问 `/product` 路由时返回 404 错误。

## 问题分析

1. **路由冲突**：在 `InitServerRoutes` 中创建了 `/product` Party，可能与直接注册的路由冲突
2. **MVC路由覆盖**：MVC路由注册可能会覆盖直接注册的路由
3. **路由注册顺序**：Iris框架中路由注册的顺序很重要

## 当前配置

### 路由注册位置
- `cmd/server/main.go`: 应用初始化
- `internal/routers/routes.go`: 路由注册逻辑

### 产品路由配置
```go
// 在 InitServerRoutes 中
app.Get("/product", middleware.AuthConProduct, func(ctx iris.Context) {
    // 处理逻辑
})

// 同时还有 Party 路由
proProduct := app.Party("/product")
proProduct.Use(middleware.AuthConProduct)
pro := mvc.New(proProduct)
pro.Handle(productController)
```

## 解决方案

### 方案1：移除Party路由（推荐）

完全移除 `/product` Party，只使用app级别的路由：

```go
// 在 InitServerRoutes 中
app.Get("/product", middleware.AuthConProduct, func(ctx iris.Context) {
    productCtrl := controllers.NewProductController(productService, orderService, sessions)
    view := productCtrl.Get(ctx)
    // ... 渲染视图
})

// 如果需要其他产品路由，使用不同的路径
app.Get("/product/detail", middleware.AuthConProduct, ...)
app.Get("/product/order", middleware.AuthConProduct, ...)
```

### 方案2：使用BeforeActivation显式注册

在ProductController中添加BeforeActivation方法：

```go
func (p *ProductController) BeforeActivation(b mvc.BeforeActivation) {
    b.Handle("GET", "/", "Get")  // 映射到 /product
    b.Handle("GET", "/detail", "GetDetail")
    b.Handle("GET", "/order", "GetOrder")
}
```

### 方案3：调整路由注册顺序

确保app级别的路由在Party路由之后注册：

```go
// 1. 先注册Party和MVC路由
proProduct := app.Party("/product")
pro := mvc.New(proProduct)
// ...

// 2. 然后在main.go中注册app级别路由（会覆盖Party路由）
app.Get("/product", ...)
```

## 推荐修复步骤

1. **检查路由注册顺序**：确保app级别路由在Party路由之后
2. **移除冲突的路由**：如果使用app级别路由，移除Party中的相同路由
3. **测试路由**：使用 `curl` 或测试程序验证路由是否正常工作

## 测试命令

```bash
# 测试产品路由
curl -H "Cookie: uid=1" http://localhost:8080/product

# 运行测试程序
go run test_login.go
```

## 注意事项

- Iris框架中，后注册的路由会覆盖先注册的路由
- Party路由和app级别路由可能产生冲突
- MVC路由的BeforeActivation可以显式控制路由映射

