package nlimit

import (
	"context"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	redis_store "github.com/ulule/limiter/v3/drivers/store/redis"
)

// Limiter 定义限流器接口
type Limiter interface {
	// Allow 检查是否允许通过
	// key: 限流键 (如 IP, UserID)
	// 返回: context (包含剩余次数等信息), error
	Allow(ctx context.Context, key string) (limiter.Context, error)
}

// limiterImpl 实现 Limiter 接口
type limiterImpl struct {
	instance *limiter.Limiter
}

// New 创建一个新的限流器
// rateStr: 限流速率字符串，格式如 "10-S" (每秒10次), "100-M" (每分钟100次), "1000-H" (每小时1000次)
// opts: 配置选项
func New(rateStr string, opts ...Option) (Limiter, error) {
	// 解析速率
	rate, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		return nil, err
	}

	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	var store limiter.Store

	if o.redisClient != nil {
		// 使用 Redis 存储
		store, err = redis_store.NewStoreWithOptions(o.redisClient, limiter.StoreOptions{
			Prefix:          o.keyPrefix,
			MaxRetry:        3,
			CleanUpInterval: time.Minute,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 使用内存存储
		store = memory.NewStore()
	}

	instance := limiter.New(store, rate)

	return &limiterImpl{instance: instance}, nil
}

func (l *limiterImpl) Allow(ctx context.Context, key string) (limiter.Context, error) {
	return l.instance.Get(ctx, key)
}
