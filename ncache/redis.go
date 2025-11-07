package ncache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"fmt"

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

// MGet 批量读取，返回命中项和值，以及未命中的 key 列表。
// 注意：不返回 TTL；如需 TTL，需要额外逐键 TTL 调用，成本较高，通常没有必要。
func (r *redisLayer) MGet(ctx context.Context, keys []string) (map[string][]byte, []string, error) {
	if len(keys) == 0 {
		return map[string][]byte{}, nil, nil
	}
	// go-redis v9: MGet 返回 []interface{}，缺失为 nil，命中多为 string 或 []byte
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, nil, err
	}
	hits := make(map[string][]byte, len(keys))
	var missing []string
	for i, v := range vals {
		if v == nil {
			missing = append(missing, keys[i])
			continue
		}
		switch t := v.(type) {
		case string:
			hits[keys[i]] = []byte(t)
		case []byte:
			hits[keys[i]] = t
		default:
			// 兜底：格式化为字符串再转字节，避免类型不匹配导致失败
			hits[keys[i]] = []byte(fmt.Sprintf("%v", t))
		}
	}
	return hits, missing, nil
}

// MSet 批量写入，使用 Pipeline 逐键 SET EX ttl。
func (r *redisLayer) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	p := r.client.Pipeline()
	for k, v := range items {
		p.Set(ctx, k, v, ttl)
	}
	_, err := p.Exec(ctx)
	return err
}
