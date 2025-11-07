package lock

import (
	"fmt"

	"github.com/go-redsync/redsync/v4"
	"github.com/jinzhu/copier"
	"github.com/samber/do"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nedis"
)

const (
	iocPrefix    = "_lock_:"
	defaultScope = "lock"
)

func Boot(scopes ...string) func() error {
	return func() error {
		if err := provide(defaultScope); err != nil {
			return fmt.Errorf("加载资源[%s]错误: %w", defaultScope, err)
		}
		for _, scope := range scopes {
			if err := provide(scope); err != nil {
				return fmt.Errorf("加载资源[%s]错误: %w", scope, err)
			}
		}
		return nil
	}
}

func Exist(scope string) bool {
	_, err := do.InvokeNamed[*redsync.Redsync](nil, iocPrefix+scope)
	return err == nil
}

func Pick(scopes ...string) *redsync.Redsync {
	scope := defaultScope
	if len(scopes) != 0 && scopes[0] != "" {
		scope = scopes[0]
	}
	return do.MustInvokeNamed[*redsync.Redsync](nil, iocPrefix+scope)
}

func provide(scope string) error {

	conf, err := getConf(scope)
	if err != nil {
		return err
	}

	redisClient := nedis.Pick()

	instance := New(redisClient, conf)

	do.ProvideNamedValue(nil, iocPrefix+scope, instance)

	return nil
}

// 获取锁配置
func GetMutex(rs *redsync.Redsync, lockName string, scopes ...string) *redsync.Mutex {
	scope := defaultScope
	if len(scopes) != 0 && scopes[0] != "" {
		scope = scopes[0]
	}

	conf, err := getConf(scope)
	if err != nil {
		return rs.NewMutex(lockName)
	}

	lockConfig, exists := conf.Locks[lockName]
	if !exists {
		lockConfig = conf.DefaultLock
	}

	return rs.NewMutex(lockName,
		redsync.WithExpiry(lockConfig.Expiry),
		redsync.WithRetryDelay(lockConfig.RetryDelay),
		redsync.WithTries(lockConfig.RetryCount+1))
}

func getConf(scope string) (conf Config, err error) {

	conf, err = defaultConfig()
	if err != nil {
		return conf, err
	}

	cfg := config.Pick()
	if !cfg.IsSet(scope) {
		return conf, fmt.Errorf("%w: 配置config.yaml[%s]不存在", kit.ErrNotFound, scope)
	}

	err = cfg.UnmarshalKey(scope, &conf)
	if err != nil {
		return conf, fmt.Errorf("%w: 解析config.yaml[%s]错误: %w", kit.ErrDataUnmarshal, scope, err)
	}
	return conf, nil
}

func defaultConfig() (conf Config, err error) {
	err = copier.CopyWithOption(&conf, &DefaultConfig, copier.Option{DeepCopy: true})
	return conf, err
}
