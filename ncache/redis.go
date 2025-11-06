package ncache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/zjutjh/mygo/nedis"
)

// 调用nedis包来获取Redis客户端
// 使用Redis实现分布式缓存层
type redisLayer struct {
	client redis.UniversalClient
}

// 创建Redis缓存层实例
func newRedisLayer(scope string) *redisLayer {
	cli := nedis.Pick(scope)
	return &redisLayer{client: cli}
}

// 实现Layer接口
func (r *redisLayer) Get(ctx context.Context, key string) ([]byte, time.Duration, bool, error) {
	res, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, 0, false, nil
		}
		return nil, 0, false, err
	}
	// 尝试获取剩余TTL
	ttl := time.Duration(0)
	d, err := r.client.TTL(ctx, key).Result()
	if err == nil && d > 0 {
		ttl = d
	}
	return res, ttl, true, nil
}

func (r *redisLayer) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, val, ttl).Err()
}

func (r *redisLayer) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
