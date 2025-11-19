package limit

import (
	"context"
	"errors"
	"time"
)

// Limiter 定义限流器通用接口
type Limiter interface {
	// Allow 判断当前请求是否被允许
	// key: 限流标识（如 user_id, ip）
	// limit: 速率（每秒允许多少次，QPS）
	// burst: 突发大小（桶容量）
	// 返回: allowed (是否允许), retryAfter (如果被拒绝，建议等待多久重试), err
	Allow(ctx context.Context, key string, limit int, burst int) (bool, time.Duration, error)
}

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)
