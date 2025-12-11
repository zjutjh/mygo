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

const UidKey = "_session_uid_"

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

// SetUid 设置uid到session
func SetUid(ctx *gin.Context, uid string) error {
	session := sessions.Default(ctx)
	session.Set(UidKey, uid)
	return session.Save()
}

// GetUid 获取session中的uid
// 注意：该函数需在 SetUid 函数进行设置后使用
func GetUid(ctx *gin.Context) (string, error) {
	session := sessions.Default(ctx)
	v := session.Get(UidKey)
	if v == nil {
		return "", fmt.Errorf("%w: 当前session中未设置[%s]", kit.ErrNotFound, UidKey)
	}
	uid, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w: 当前session设置[%s]类型错误", kit.ErrDataFormat, UidKey)
	}
	return uid, nil
}

// DeleteUid 删除session中的uid
func DeleteUid(ctx *gin.Context) error {
	session := sessions.Default(ctx)
	session.Delete(UidKey)
	return session.Save()
}
