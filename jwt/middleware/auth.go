package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/zjutjh/mygo/foundation/reply"
	myjwt "github.com/zjutjh/mygo/jwt"
	"github.com/zjutjh/mygo/kit"
)

// Auth JWT 鉴权中间件
// 参数: mustLogged 是否必须登录
func Auth[T any](mustLogged bool, scopes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 Token
		token := ctx.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			if !mustLogged {
				ctx.Next()
				return
			}
			reply.Fail(ctx, kit.CodeNotLoggedIn)
			return
		}

		// 解析 Token
		token = token[7:]
		identity, err := myjwt.Pick[T](scopes...).ParseToken(token)
		if err != nil {
			if !mustLogged {
				ctx.Next()
				return
			}
			// Token 过期
			if errors.Is(err, jwt.ErrTokenExpired) {
				reply.Fail(ctx, kit.CodeLoginExpired)
				return
			}
			reply.Fail(ctx, kit.CodeDataParseError)
			return
		}

		// 挂载 identity
		myjwt.MountIdentity(ctx, identity)
	}
}
