package jwt

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/kit"
)

const MountKey = "_jwt_identity_"

// MountIdentity 挂载 identity 至上下文
func MountIdentity[T any](ctx *gin.Context, identity T) {
	ctx.Set(MountKey, identity)
}

// GetIdentity 获取上下文中挂载的 identity
// 注意：该函数需在 MountIdentity 函数进行挂载后使用
func GetIdentity[T any](ctx *gin.Context) (T, error) {
	var zero T
	v, ok := ctx.Get(MountKey)
	if !ok {
		return zero, fmt.Errorf("%w: 当前上下文未挂载[%s]", kit.ErrNotFound, MountKey)
	}
	identity, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("%w: 当前上下文挂载[%s]类型错误", kit.ErrDataFormat, MountKey)
	}
	return identity, nil
}
