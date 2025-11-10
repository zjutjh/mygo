package miniprogram

import (
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/jinzhu/copier"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/config"
)

const (
	iocPrefix    = "_wechat_mini_program_"
	defaultScope = "wechat_mini_program"
)

// Boot 预加载默认实例
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

// Exist 判断实例是否挂载（被Boot过）且类型正确
func Exist(scope string) bool {
	_, err := do.InvokeNamed[*miniProgram.MiniProgram](nil, iocPrefix+scope)
	return err == nil
}

// Pick 获取指定scope实例
func Pick(scopes ...string) *miniProgram.MiniProgram {
	scope := defaultScope
	if len(scopes) != 0 && scopes[0] != "" {
		scope = scopes[0]
	}
	return do.MustInvokeNamed[*miniProgram.MiniProgram](nil, iocPrefix+scope)
}

// provide 提供指定scope实例
func provide(scope string) error {
	conf, err := getConf(scope)
	if err != nil {
		return err
	}
	// 初始化实例
	instance, _ := New(conf)
	// 挂载实例
	do.ProvideNamedValue(nil, iocPrefix+scope, instance)
	return nil
}

func getConf(scope string) (conf Config, err error) {
	// 初始化默认配置
	conf, err = defaultConfig()
	if err != nil {
		return conf, err
	}
	//判断 scope 配置是否存在
	cfg := config.Pick()
	if !cfg.IsSet(scope) {
		return conf, fmt.Errorf("配置[%s]不存在", scope)
	}
	//解析配置
	if err := cfg.UnmarshalKey(scope, &conf); err != nil {
		return conf, fmt.Errorf("解析配置[%s]错误: %w", scope, err)
	}
	return conf, nil
}

// defaultConfig 获取默认配置
func defaultConfig() (conf Config, err error) {
	err = copier.CopyWithOption(&conf, &DefaultConfig, copier.Option{DeepCopy: true})
	return conf, err
}
