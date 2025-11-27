package ncache

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

type User struct {
	ID   int
	Name string
}

func TestCache_Memory(t *testing.T) {
	// 创建纯内存缓存实例
	c := New()

	ctx := context.Background()
	key := "user:1"
	user := &User{ID: 1, Name: "Hastune Miku"}

	// Set
	err := c.Set(ctx, key, user)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get
	var gotUser User
	_, err = c.Get(ctx, key, &gotUser)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if gotUser.ID != 1 || gotUser.Name != "Hastune Miku" {
		t.Errorf("Expected user {1, Hastune Miku}, got %v", gotUser)
	}
}

func TestCache_RedisChain(t *testing.T) {
	// 尝试连接本地 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis寄B了，跳过: %v", err)
	}

	// 清理测试数据
	testKey := "test:chain:user:1"
	rdb.Del(ctx, testKey)

	// 初始化多级缓存 (Memory + Redis)
	c := New(WithRedis(rdb))

	user := &User{ID: 99, Name: "Miku"}

	// 测试写入 (Set) - 应该同时写入 Memory 和 Redis
	if err := c.Set(ctx, testKey, user); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 验证 Redis 里是否有数据
	if exists := rdb.Exists(ctx, testKey).Val(); exists == 0 {
		t.Error("Key should exist in Redis after Set")
	}

	// 测试读取 (Get)
	var gotUser User
	if _, err := c.Get(ctx, testKey, &gotUser); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if gotUser.Name != "Miku" {
		t.Errorf("Expected Miku, got %s", gotUser.Name)
	}

	// 测试回填
	// 手动删除 Memory 里的数据 (模拟过期或重启)，但保留 Redis 里的数据
	// 通过创建一个新的 CacheManager 来模拟“重启”
	// 新的 CacheManager 的 Memory 是空的，但 Redis 还是同一个
	c2 := New(WithRedis(rdb))

	var gotUser2 User
	// L1 Miss -> L2 Hit -> Backfill L1
	if _, err := c2.Get(ctx, testKey, &gotUser2); err != nil {
		t.Fatalf("Get from c2 failed: %v", err)
	}
	if gotUser2.Name != "Miku" {
		t.Errorf("Expected Miku from L2, got %s", gotUser2.Name)
	}

	// 清理
	c.Delete(ctx, testKey)
}
