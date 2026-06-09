package ncache

import (
	"github.com/samber/do"
)

const (
	iocPrefix    = "_cache_:"
	defaultScope = "cache"
)

// Boot 默认引导器，注册一个或多个内存缓存实例
func Boot(scopes ...string) func() error {
	return func() error {
		// 如果未指定 scope，则使用默认的 scope
		if len(scopes) == 0 {
			scopes = []string{defaultScope}
		}

		// 按 scope 注册内存缓存实例
		for _, scope := range scopes {
			if scope == "" {
				scope = defaultScope
			}
			do.ProvideNamedValue(nil, iocPrefix+scope, New())
		}

		return nil
	}
}

// Pick 获取缓存实例
func Pick(scopes ...string) CacheManager {
	scope := defaultScope
	if len(scopes) > 0 && scopes[0] != "" {
		scope = scopes[0]
	}
	return do.MustInvokeNamed[CacheManager](nil, iocPrefix+scope)
}
