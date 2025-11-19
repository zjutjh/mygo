package limit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// getRedisClient 获取本地 Redis 连接，如果连不上则跳过测试
func getRedisClient(t *testing.T) redis.UniversalClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("skipping test: redis not available: %v", err)
	}
	return rdb
}

func TestTokenBucketLimiter(t *testing.T) {
	rdb := getRedisClient(t)
	limiter := NewTokenBucketLimiter(rdb)
	ctx := context.Background()
	key := "test_token_bucket"

	// 清理旧数据
	rdb.Del(ctx, key)

	// 配置: 速率 1次/秒, 桶容量 2
	limit := 1
	burst := 2

	t.Logf("Testing TokenBucket: Rate=%d, Burst=%d", limit, burst)

	// 1. 第1次请求 (应该允许)
	allowed, _, err := limiter.Allow(ctx, key, limit, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected 1st request to be allowed")
	} else {
		t.Log("1st request allowed")
	}

	// 2. 第2次请求 (应该允许 - 消耗突发容量)
	allowed, _, err = limiter.Allow(ctx, key, limit, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected 2nd request to be allowed")
	} else {
		t.Log("2nd request allowed")
	}

	// 3. 第3次请求 (应该拒绝 - 桶空了)
	allowed, retry, err := limiter.Allow(ctx, key, limit, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected 3rd request to be rejected")
	} else {
		t.Logf("3rd request rejected, retry after: %v", retry)
	}

	// 4. 等待并重试
	if retry > 0 {
		t.Logf("Waiting for %v...", retry)
		time.Sleep(retry + 100*time.Millisecond)
		allowed, _, err = limiter.Allow(ctx, key, limit, burst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Error("expected request after wait to be allowed")
		} else {
			t.Log("Request after wait allowed")
		}
	}
}

func TestLeakyBucketLimiter(t *testing.T) {
	rdb := getRedisClient(t)
	limiter := NewLeakyBucketLimiter(rdb)
	ctx := context.Background()
	key := "test_leaky_bucket"

	// 清理旧数据
	rdb.Del(ctx, key)

	// 配置: 速率 10次/秒 (即每100ms允许1次), 突发 1 (严格平滑)
	limit := 10
	burst := 1

	t.Logf("Testing LeakyBucket (GCRA): Rate=%d, Burst=%d", limit, burst)

	// 1. 第1次请求 (应该允许)
	allowed, _, err := limiter.Allow(ctx, key, limit, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected 1st request to be allowed")
	} else {
		t.Log("1st request allowed")
	}

	// 2. 立即发起第2次请求 (应该拒绝 - 因为间隔未到且burst=1)
	allowed, retry, err := limiter.Allow(ctx, key, limit, burst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected 2nd immediate request to be rejected with burst=1")
	} else {
		t.Logf("2nd immediate request rejected, retry after: %v", retry)
	}
}
