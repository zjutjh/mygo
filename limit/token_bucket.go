package limit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	rdb redis.UniversalClient
}

func NewTokenBucketLimiter(rdb redis.UniversalClient) *TokenBucketLimiter {
	return &TokenBucketLimiter{rdb: rdb}
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string, limit int, burst int) (bool, time.Duration, error) {
	if limit <= 0 {
		return false, 0, nil
	}
	if burst <= 0 {
		burst = limit
	}

	now := time.Now().Unix() // 秒级
	cost := 1

	// 执行脚本
	res, err := l.rdb.Eval(ctx, scriptTokenBucket, []string{key}, burst, limit, now, cost).Result()
	if err != nil {
		return false, 0, err
	}

	// 解析结果 {allowed, retry_after}
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return false, 0, nil
	}

	allowed := arr[0].(int64) == 1
	retryAfter := time.Duration(arr[1].(int64)) * time.Second

	return allowed, retryAfter, nil
}
