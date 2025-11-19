package ncache

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

func TestRedisLayer_Basic(t *testing.T) {
	rdb := getRedisClient(t)
	// 手动构造 redisLayer，绕过 nedis 依赖
	layer := &redisLayer{client: rdb}
	ctx := context.Background()
	key := "test_ncache_basic"
	val := []byte("hello world")

	// Clean up
	rdb.Del(ctx, key)

	// 1. Set
	err := layer.Set(ctx, key, val, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 2. Get
	got, ttl, found, err := layer.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Error("Get expected found=true, got false")
	}
	if string(got) != string(val) {
		t.Errorf("Get expected %s, got %s", val, got)
	}
	if ttl <= 0 {
		t.Errorf("Get expected ttl > 0, got %v", ttl)
	}

	// 3. Delete
	err = layer.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 4. Get after Delete
	_, _, found, err = layer.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if found {
		t.Error("Get after delete expected found=false, got true")
	}
}

func TestRedisLayer_Batch(t *testing.T) {
	rdb := getRedisClient(t)
	layer := &redisLayer{client: rdb}
	ctx := context.Background()

	items := map[string][]byte{
		"k1": []byte("v1"),
		"k2": []byte("v2"),
		"k3": []byte("v3"),
	}

	// Clean up
	for k := range items {
		rdb.Del(ctx, k)
	}

	// 1. MSet
	err := layer.MSet(ctx, items, time.Minute)
	if err != nil {
		t.Fatalf("MSet failed: %v", err)
	}

	// 2. MGet (All hits)
	keys := []string{"k1", "k2", "k3"}
	hits, missing, err := layer.MGet(ctx, keys)
	if err != nil {
		t.Fatalf("MGet failed: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("MGet expected 0 missing, got %v", missing)
	}
	if len(hits) != 3 {
		t.Errorf("MGet expected 3 hits, got %d", len(hits))
	}
	for k, v := range items {
		if string(hits[k]) != string(v) {
			t.Errorf("MGet key %s expected %s, got %s", k, v, hits[k])
		}
	}

	// 3. MGet (Partial hits)
	keys = []string{"k1", "k_missing", "k3"}
	hits, missing, err = layer.MGet(ctx, keys)
	if err != nil {
		t.Fatalf("MGet partial failed: %v", err)
	}
	if len(missing) != 1 || missing[0] != "k_missing" {
		t.Errorf("MGet partial expected missing [k_missing], got %v", missing)
	}
	if len(hits) != 2 {
		t.Errorf("MGet partial expected 2 hits, got %d", len(hits))
	}
}
