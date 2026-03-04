package nlimit

import (
	"github.com/samber/do"
)

const (
	iocPrefix    = "_limit_:"
	defaultScope = "limit"
)

// Boot 注册默认限流器 (内存模式, 默认 100 req/s)
// 你可以在应用启动时根据配置覆盖这个默认值
func Boot() func() error {
	return func() error {
		// 默认注册一个内存限流器，防止未配置时报错
		l, err := New("100-S")
		if err != nil {
			return err
		}
		do.ProvideNamedValue(nil, iocPrefix+defaultScope, l)
		return nil
	}
}

// Pick 获取指定 scope 的限流器
func Pick(scope string) Limiter {
	if scope == "" {
		scope = defaultScope
	}
	return do.MustInvokeNamed[Limiter](nil, iocPrefix+scope)
}
