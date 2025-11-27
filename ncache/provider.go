package ncache

import (
	"github.com/samber/do"
)

const (
	iocPrefix    = "_cache_:"
	defaultScope = "cache"
)

// Boot 默认引导器，注册默认的内存缓存实例
func Boot() func() error {
	return func() error {
		// 注册默认的内存缓存实例
		do.ProvideNamedValue(nil, iocPrefix+defaultScope, New())
		return nil
	}
}

// Pick 获取缓存实例
func Pick(scope string) CacheManager {
	if scope == "" {
		scope = defaultScope
	}
	return do.MustInvokeNamed[CacheManager](nil, iocPrefix+scope)
}
