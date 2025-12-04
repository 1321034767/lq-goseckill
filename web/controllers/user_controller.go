package controllers

import (
	"net/http"
	"time"

	"github.com/kataras/iris/v12"

	"github.com/example/goseckill/internal/service"
)

// UserController 负责前台登录/注册页面与表单处理。
type UserController struct {
	userService *service.UserService
}

// NewUserController 构造函数，供路由层复用同一套逻辑。
func NewUserController(userSvc *service.UserService) *UserController {
	return &UserController{userService: userSvc}
}

// ShowLogin 渲染登录表单。
func (c *UserController) ShowLogin(ctx iris.Context) {
	if err := ctx.View("user/login.html"); err != nil {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>无法加载登录页面，请稍后重试。</h2>")
	}
}

// ShowRegister 渲染注册表单。
func (c *UserController) ShowRegister(ctx iris.Context) {
	if err := ctx.View("user/register.html"); err != nil {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>无法加载注册页面，请稍后重试。</h2>")
	}
}

// ShowManage 简单的用户中心占位页。
func (c *UserController) ShowManage(ctx iris.Context) {
	if err := ctx.View("user/manage.html"); err != nil {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>无法加载用户中心，请稍后重试。</h2>")
	}
}

// PostLogin 处理登录表单提交，成功后写 cookie 并跳回首页。
func (c *UserController) PostLogin(ctx iris.Context) {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	if username == "" || password == "" {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>用户名和密码不能为空</h2>")
		return
	}

	token, err := c.userService.Login(ctx.Request().Context(), username, password)
	if err != nil {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>登录失败: " + err.Error() + "</h2>")
		return
	}

	ctx.SetCookie(&http.Cookie{
		Name:  "username",
		Value: username,
		Path:  "/",
	})
	ctx.SetCookie(&http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	ctx.Redirect("/", iris.StatusFound)
}

// PostAdd 处理注册表单提交，成功后跳转到登录页。
func (c *UserController) PostAdd(ctx iris.Context) {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	if username == "" || password == "" {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>用户名和密码不能为空</h2>")
		return
	}

	_, err := c.userService.Register(ctx.Request().Context(), username, password)
	if err != nil {
		ctx.ContentType("text/html; charset=utf-8")
		_, _ = ctx.WriteString("<h2>注册失败: " + err.Error() + "</h2>")
		return
	}

	ctx.Redirect("/login", iris.StatusFound)
}

// Logout 清理 cookie 并回到首页。
func (c *UserController) Logout(ctx iris.Context) {
	clearCookie := func(name string) {
		ctx.SetCookie(&http.Cookie{
			Name:    name,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			MaxAge:  -1,
		})
	}
	clearCookie("username")
	clearCookie("token")
	ctx.Redirect("/", iris.StatusFound)
}
