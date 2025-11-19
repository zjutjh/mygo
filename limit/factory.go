package limit

import (
	"fmt"
	"strings"

	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/nlog"
)

const (
	TypeTokenBucket = "token_bucket"
	TypeLeakyBucket = "leaky_bucket"
)

// NewLimiter 创建一个限流器实例
// redisScope: nedis 中配置的 scope 名称 (如 "default", "rate_limit")
// algoType: 限流算法类型 (token_bucket / leaky_bucket)
func NewLimiter(redisScope string, algoType string) (Limiter, error) {
	if !nedis.Exist(redisScope) {
		return nil, fmt.Errorf("limit: 依赖的 redis scope [%s] 未找到，请确保在调用 NewLimiter 之前已在 BootList 中执行 nedis.Boot() 且配置正确", redisScope)
	}
	rdb := nedis.Pick(redisScope)

	switch strings.ToLower(algoType) {
	case TypeTokenBucket:
		return NewTokenBucketLimiter(rdb), nil
	case TypeLeakyBucket:
		return NewLeakyBucketLimiter(rdb), nil
	default:
		// 默认使用令牌桶
		return NewTokenBucketLimiter(rdb), nil
	}
}

// NewLimiterWithLog 创建一个带日志记录的限流器实例
// logScope: nlog 中配置的 scope 名称
func NewLimiterWithLog(redisScope string, algoType string, logScope string) (Limiter, error) {
	l, err := NewLimiter(redisScope, algoType)
	if err != nil {
		return nil, err
	}

	if !nlog.Exist(logScope) {
		return nil, fmt.Errorf("limit: 依赖的 log scope [%s] 未找到，请确保在调用 NewLimiterWithLog 之前已在 BootList 中执行 nlog.Boot() 且配置正确", logScope)
	}

	return &loggingLimiter{
		next:   l,
		logger: nlog.Pick(logScope),
	}, nil
}
