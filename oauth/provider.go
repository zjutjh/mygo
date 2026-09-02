package oauth

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/samber/do"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/kit"
)

const (
	iocPrefix    = "_oauth_:"
	defaultScope = "oauth"
)

// Boot 预加载默认实例，同时加载指定实例列表
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

// Exist 判断 scope 实例是否已挂载且类型正确
func Exist(scopes ...string) bool {
	scope := defaultScope
	if len(scopes) != 0 && scopes[0] != "" {
		scope = scopes[0]
	}
	_, err := do.InvokeNamed[*OAuth](nil, iocPrefix+scope)
	return err == nil
}

// Pick 获取指定 scope 的 OAuth 实例
func Pick(scopes ...string) *OAuth {
	scope := defaultScope
	if len(scopes) != 0 && scopes[0] != "" {
		scope = scopes[0]
	}
	return do.MustInvokeNamed[*OAuth](nil, iocPrefix+scope)
}

func provide(scope string) error {
	conf, err := getConf(scope)
	if err != nil {
		return err
	}

	instance := New(conf)
	do.ProvideNamedValue(nil, iocPrefix+scope, instance)
	return nil
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
	if err = cfg.UnmarshalKey(scope, &conf); err != nil {
		return conf, fmt.Errorf("%w: 解析config.yaml[%s]错误: %w", kit.ErrDataUnmarshal, scope, err)
	}
	return conf, nil
}

func defaultConfig() (conf Config, err error) {
	err = copier.CopyWithOption(&conf, &DefaultConfig, copier.Option{DeepCopy: true})
	return conf, err
}
