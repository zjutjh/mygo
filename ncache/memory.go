package ncache

import (
	"context"
	"sync"
	"time"
)

// 使用sync.Map实现内存缓存层
type memoryItem struct {
	v        []byte
	expireAt time.Time // 零值表示永不过期
}

// 定义主存储层
type memoryLayer struct {
	m       sync.Map
	max     int
	ticker  *time.Ticker
	closing chan struct{}
}

func newMemoryLayer(conf MemoryConfig) *memoryLayer {
	ml := &memoryLayer{
		max:     conf.MaxEntries,
		closing: make(chan struct{}),
	}
	// 如果配置了 conf.CleanInterval，则启动定期清理过期缓存的协程
	if conf.CleanInterval > 0 {
		// 启动一个 janitor Goroutine 定期清理过期缓存
		ml.ticker = time.NewTicker(conf.CleanInterval)
		go ml.janitor()
	}
	return ml
}

// 过期缓存清理机制
func (m *memoryLayer) janitor() {
	for {
		select {
		case <-m.closing:
			return
		case <-m.ticker.C:
			now := time.Now()
			// 粗略清理（遍历 map）
			m.m.Range(func(key, value any) bool {
				it, _ := value.(memoryItem)
				if !it.expireAt.IsZero() && now.After(it.expireAt) {
					m.m.Delete(key)
				}
				return true
			})
		}
	}
}

func (m *memoryLayer) stop() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.closing)
}

func (m *memoryLayer) Get(ctx context.Context, key string) ([]byte, time.Duration, bool, error) {
	//首先尝试用 m.m.Load(key) 加载值
	v, ok := m.m.Load(key)
	if !ok {
		return nil, 0, false, nil
	}
	it := v.(memoryItem)
	// 检查缓存是否过期
	if !it.expireAt.IsZero() && time.Now().After(it.expireAt) {
		m.m.Delete(key)
		return nil, 0, false, nil
	}
	var ttlLeft time.Duration
	if !it.expireAt.IsZero() {
		ttlLeft = time.Until(it.expireAt)
	}
	return it.v, ttlLeft, true, nil
}

func (m *memoryLayer) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	m.m.Store(key, memoryItem{v: val, expireAt: expireAt})
	// 简单数量控制（软限制）
	if m.max > 0 {
		count := 0
		m.m.Range(func(_, _ any) bool { count++; return count <= m.max })
		if count > m.max {
			// 触发一次清理以回到软上限（不保证严格）
			now := time.Now()
			m.m.Range(func(key, value any) bool {
				it, _ := value.(memoryItem)
				if !it.expireAt.IsZero() && now.After(it.expireAt) {
					m.m.Delete(key)
				}
				return true
			})
		}
	}
	return nil
}

func (m *memoryLayer) Delete(ctx context.Context, key string) error {
	m.m.Delete(key)
	return nil
}
