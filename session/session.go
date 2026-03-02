package session

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/session/redis"
)

const defaultConfigKey = "session"

const IdentityKey = "_session_identity_"

// Pick 获取指定实例
func Pick(keys ...string) gin.HandlerFunc {
	key := defaultConfigKey
	if len(keys) != 0 && keys[0] != "" {
		key = keys[0]
	}
	conf := Config{}
	err := copier.Copy(&conf, DefaultConfig)
	if err != nil {
		panic(err)
	}
	app := config.Pick()
	if !app.IsSet(key) {
		panic(kit.ErrNotFound)
	}
	err = app.UnmarshalKey(key, &conf)
	if err != nil {
		panic(err)
	}

	var store sessions.Store
	keyPairs := []byte(conf.Secret)
	switch conf.Driver {
	case DriverRedis:
		var err error
		rdb := nedis.Pick(conf.Redis)
		store, err = redis.NewStore(rdb, keyPairs)
		if err != nil {
			panic(err)
		}
	case DriverMemory:
		store = memstore.NewStore(keyPairs)
	default:
		store = memstore.NewStore(keyPairs)
	}
	store.Options(sessions.Options{
		Path:     conf.Path,
		Domain:   conf.Domain,
		MaxAge:   conf.MaxAge,
		Secure:   conf.Secure,
		HttpOnly: conf.HttpOnly,
		SameSite: conf.SameSite,
	})

	return sessions.Sessions(conf.Name, store)
}

// SetIdentity 设置 identity 到session
func SetIdentity[T any](ctx *gin.Context, identity T) error {
	session := sessions.Default(ctx)
	session.Set(IdentityKey, identity)
	return session.Save()
}

// GetIdentity 获取 session 中的 identity
// 注意：该函数需在 SetIdentity 函数进行设置后使用
func GetIdentity[T any](ctx *gin.Context) (T, error) {
	var zero T
	session := sessions.Default(ctx)
	v := session.Get(IdentityKey)
	if v == nil {
		return zero, fmt.Errorf("%w: 当前session中未设置[%s]", kit.ErrNotFound, IdentityKey)
	}
	identity, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("%w: 当前session设置[%s]类型错误", kit.ErrDataFormat, IdentityKey)
	}
	return identity, nil
}

// DeleteIdentity 删除 session 中的 identity
func DeleteIdentity(ctx *gin.Context) error {
	session := sessions.Default(ctx)
	session.Delete(IdentityKey)
	return session.Save()
}
