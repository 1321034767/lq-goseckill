package controllers

import (
	"GoSecKill/internal/services"
	"GoSecKill/pkg/models"
	"GoSecKill/pkg/util"
	"net/http"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	sessions    *sessions.Sessions
	userService services.IUserService
}

func NewUserController(userService services.IUserService, sessions *sessions.Sessions) *UserController {
	return &UserController{userService: userService, sessions: sessions}
}

func (c *UserController) GetUserList() mvc.View {
	users, _ := c.userService.GetUserList()
	zap.L().Info("Get user list", zap.Any("users", users))
	return mvc.View{
		Name: "user/view.html",
		Data: iris.Map{
			"users": users,
		},
	}
}

func (c *UserController) PostUpdate(ctx iris.Context) {
	user := models.User{}

	if err := ctx.ReadForm(&user); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}

	err := c.userService.UpdateUser(user)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to update user")
		return
	}

	zap.L().Info("Successfully updated user", zap.Any("user", user))
	ctx.Redirect("/user/all")
}

func (c *UserController) GetAddUser(ctx iris.Context) mvc.View {
	return mvc.View{
		Name: "user/add.html",
	}
}

func (c *UserController) PostRegister(ctx iris.Context) {
	user := models.User{}
	var err error

	if err := ctx.ReadForm(&user); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to read form")
		return
	}
	user.Password, err = encryptUserPassword(user.Password)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to encrypt password")
		return
	}

	err = c.userService.InsertUser(user)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to insert user")
		return
	}

	zap.L().Info("Successfully inserted user", zap.Any("user", user))
	ctx.Redirect("/user/login")
}

func (c *UserController) PostLogin(ctx iris.Context) {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	users, _ := c.userService.GetUserListByUsername(username)
	if len(users) == 0 || bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(password)) != nil {
		ctx.StatusCode(iris.StatusUnauthorized)
		_, _ = ctx.WriteString("Wrong username or password")
		return
	}

	ctx.SetCookie(&http.Cookie{
		Name:  "uid",
		Value: strconv.FormatInt(int64(users[0].ID), 10),
		Path:  "/",
	})
	uidByte := []byte(strconv.FormatUint(uint64(users[0].ID), 10))
	uidStr, _ := util.Encrypt(uidByte)
	ctx.SetCookie(&http.Cookie{
		Name:  "sign",
		Value: string(uidStr),
		Path:  "/",
	})
	c.sessions.Start(ctx).Set("userID", strconv.FormatInt(int64(users[0].ID), 10))

	zap.L().Info("User login", zap.String("username", username))
	ctx.Redirect("/product")
}

func (c *UserController) GetDelete(ctx iris.Context) {
	id, _ := ctx.Params().GetInt("id")
	err := c.userService.DeleteUser(id)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_, _ = ctx.WriteString("Failed to delete user")
		return
	}

	zap.L().Info("Successfully deleted user", zap.Int("id", id))
	ctx.Redirect("/user/all")
}

func (c *UserController) GetRegister(ctx iris.Context) mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}

func (c *UserController) GetLogin(ctx iris.Context) mvc.View {
	return mvc.View{
		Name: "user/login.html",
	}
}

func encryptUserPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (c *UserController) BeforeActivation(b mvc.BeforeActivation) {
	//b.Handle("GET", "/all", "GetUserList")
	//b.Handle("POST", "/update", "PostUpdate")
	b.Handle("GET", "/add", "GetAddUser")
	b.Handle("POST", "/add", "PostRegister")
	b.Handle("POST", "/login", "PostLogin")
	b.Handle("GET", "/register", "GetRegister")
	b.Handle("GET", "/login", "GetLogin")
	//b.Handle("GET", "/delete/{id:int}", "GetDelete")
}
