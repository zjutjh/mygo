package limit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LeakyBucketLimiter struct {
	rdb redis.UniversalClient
}

func NewLeakyBucketLimiter(rdb redis.UniversalClient) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{rdb: rdb}
}

func (l *LeakyBucketLimiter) Allow(ctx context.Context, key string, limit int, burst int) (bool, time.Duration, error) {
	if limit <= 0 {
		return false, 0, nil
	}
	// GCRA 中 burst 0 意味着没有任何突发容忍，严格按间隔
	// 但通常至少给 1
	if burst < 1 {
		burst = 1
	}

	now := time.Now().UnixMilli() // 毫秒级

	res, err := l.rdb.Eval(ctx, scriptGCRA, []string{key}, limit, burst, now).Result()
	if err != nil {
		return false, 0, err
	}

	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, 0, nil
	}

	allowed := arr[0].(int64) == 1
	retryAfter := time.Duration(arr[1].(int64)) * time.Second

	return allowed, retryAfter, nil
}
