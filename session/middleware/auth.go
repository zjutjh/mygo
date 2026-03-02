package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/session"
)

// Auth Session 鉴权中间件
// 参数: mustLogged 是否必须登录
func Auth[T any](mustLogged bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := session.GetIdentity[T](ctx)
		if err != nil {
			if !mustLogged {
				ctx.Next()
				return
			}
			reply.Fail(ctx, kit.CodeNotLoggedIn)
			return
		}
	}
}
