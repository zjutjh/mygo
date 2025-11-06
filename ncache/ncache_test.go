package ncache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testConfOnlyMemory() Config {
	return Config{
		Enable:        true,
		Prefix:        "",
		DefaultTTL:    200 * time.Millisecond,
		JitterPercent: 0, // 测试中避免随机性
		NegativeTTL:   80 * time.Millisecond,
		Policy:        string(PolicyCacheAsideSingleFlight),
		Layers: []LayerConfig{
			{Type: "memory", Memory: MemoryConfig{MaxEntries: 1000, CleanInterval: 0}},
		},
	}
}

func TestNew_WithRedisLayerWithoutBoot_ReturnsError(t *testing.T) {
	conf := Config{
		Enable:        true,
		DefaultTTL:    1 * time.Second,
		JitterPercent: 0,
		Layers: []LayerConfig{
			{Type: "redis", Redis: RedisConfig{Scope: "redis"}},
		},
	}
	_, err := New(conf)
	if err == nil {
		t.Fatalf("expected error when redis layer configured without nedis.Boot, got nil")
	}
}

func TestSetGet_MemoryOnly(t *testing.T) {
	conf := testConfOnlyMemory()
	c, err := New(conf)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	if cl, ok := c.(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = cl.Close() })
	}
	ctx := context.Background()
	key := "t:setget"
	val := []byte("hello")
	if err := c.Set(ctx, key, val, 50*time.Millisecond); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != string(val) {
		t.Fatalf("value mismatch: got=%s want=%s", string(got), string(val))
	}
}

func TestRemember_SingleFlight_OnlyOnce(t *testing.T) {
	conf := testConfOnlyMemory()
	c, err := New(conf)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	if cl, ok := c.(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = cl.Close() })
	}
	ctx := context.Background()
	key := "t:sf:onlyonce"
	var calls int32
	loader := func(ctx context.Context) ([]byte, time.Duration, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(30 * time.Millisecond) // 放大并发窗口
		return []byte("v"), 100 * time.Millisecond, nil
	}
	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := c.Remember(ctx, key, loader)
			if err != nil && err != ErrNotFound {
				t.Errorf("remember: %v", err)
			}
		}()
	}
	wg.Wait()
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("loader called %d times, want 1", calls)
	}
}

func TestRemember_NegativeCache(t *testing.T) {
	conf := testConfOnlyMemory()
	conf.NegativeTTL = 80 * time.Millisecond
	conf.JitterPercent = 0
	c, err := New(conf)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	if cl, ok := c.(interface{ Close() error }); ok {
		t.Cleanup(func() { _ = cl.Close() })
	}
	ctx := context.Background()
	key := "t:neg"
	var calls int32
	loader := func(ctx context.Context) ([]byte, time.Duration, error) {
		atomic.AddInt32(&calls, 1)
		return nil, 0, ErrNotFound
	}
	if _, err := c.Remember(ctx, key, loader); err != ErrNotFound { // 写入负缓存
		t.Fatalf("first remember want ErrNotFound, got %v", err)
	}
	// 第二次直接 Get，应命中负缓存并返回 ErrNotFound，且不触发 loader
	if _, err := c.Get(ctx, key); err != ErrNotFound {
		t.Fatalf("get after negative should return ErrNotFound, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("loader called %d times, want 1", atomic.LoadInt32(&calls))
	}
	time.Sleep(conf.NegativeTTL + 20*time.Millisecond)
	_, _ = c.Remember(ctx, key, loader)
	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("loader should be called again after negative TTL expires")
	}
}

func TestApplyJitterBounds(t *testing.T) {
	// 使用较大的 base 以避免被 clamp 到 1s 下限导致测试失败
	c := &multiCache{conf: Config{JitterPercent: 0.2}}
	base := 5 * time.Second
	min := time.Duration(float64(base) * 0.8)
	max := time.Duration(float64(base) * 1.2)
	for i := 0; i < 300; i++ {
		adj := c.applyJitter(base)
		if adj < min || adj > max {
			t.Fatalf("jitter out of bounds: %v not in [%v,%v]", adj, min, max)
		}
	}
}
