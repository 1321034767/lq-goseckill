package middleware

import (
	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
)

func AuthConProduct(ctx iris.Context) {
	uid := ctx.GetCookie("uid")
	if uid == "" {
		zap.L().Error("uid is empty")
		ctx.Redirect("/user/login")
		return
	}
	zap.L().Debug("uid", zap.String("uid", uid))
	ctx.Next()
}
