package ncache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zjutjh/mygo/nlog"
)

// 使用sync.Map实现内存缓存层
type memoryItem struct {
	v        []byte
	expireAt time.Time // 零值表示永不过期
}

// 内存缓存层（L1）。简单实现：软数量上限 + 过期清理 + 可选命中率统计。
type memoryLayer struct {
	m       sync.Map
	max     int
	ticker  *time.Ticker // 过期清理
	closing chan struct{}

	// 命中统计
	gets uint64
	hits uint64

	// 命中率日志
	statsTicker   *time.Ticker
	statsLogScope string
}

func newMemoryLayer(conf MemoryConfig) *memoryLayer {
	ml := &memoryLayer{
		max:     conf.MaxEntries,
		closing: make(chan struct{}),
	}
	if conf.CleanInterval > 0 {
		ml.ticker = time.NewTicker(conf.CleanInterval)
		go ml.janitor()
	}
	if conf.StatsInterval > 0 {
		ml.statsTicker = time.NewTicker(conf.StatsInterval)
		ml.statsLogScope = conf.StatsLog
		go ml.statsLogger()
	}
	return ml
}

// 定期清理过期条目（粗粒度遍历）
func (m *memoryLayer) janitor() {
	for {
		select {
		case <-m.closing:
			return
		case <-m.ticker.C:
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
}

// Close 关闭后台定时器与协程
func (m *memoryLayer) Close() error {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	if m.statsTicker != nil {
		m.statsTicker.Stop()
	}
	close(m.closing)
	return nil
}

func (m *memoryLayer) Get(ctx context.Context, key string) ([]byte, time.Duration, bool, error) {
	atomic.AddUint64(&m.gets, 1)
	v, ok := m.m.Load(key)
	if !ok {
		return nil, 0, false, nil
	}
	it := v.(memoryItem)
	if !it.expireAt.IsZero() && time.Now().After(it.expireAt) {
		m.m.Delete(key)
		return nil, 0, false, nil
	}
	var ttlLeft time.Duration
	if !it.expireAt.IsZero() {
		ttlLeft = time.Until(it.expireAt)
	}
	atomic.AddUint64(&m.hits, 1)
	return it.v, ttlLeft, true, nil
}

func (m *memoryLayer) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	m.m.Store(key, memoryItem{v: val, expireAt: expireAt})

	// 软上限：超过 maxEntries 后触发一次过期清理（不保证严格约束）
	if m.max > 0 {
		count := 0
		m.m.Range(func(_, _ any) bool { count++; return count <= m.max+1 })
		if count > m.max {
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

// 周期性输出命中率（命中率 = hits / gets）。仅在配置了 StatsInterval 时开启。
func (m *memoryLayer) statsLogger() {
	var logger *logrus.Logger
	pick := func() *logrus.Logger {
		if m.statsLogScope != "" {
			if nlog.Exist(m.statsLogScope) {
				return nlog.Pick(m.statsLogScope)
			}
			return nil
		}
		if nlog.Exist("log") {
			return nlog.Pick("log")
		}
		return nil
	}
	logger = pick()
	for {
		select {
		case <-m.closing:
			return
		case <-m.statsTicker.C:
			if logger == nil { // 应用可能稍后 Boot 日志
				logger = pick()
				if logger == nil {
					continue
				}
			}
			g := atomic.LoadUint64(&m.gets)
			h := atomic.LoadUint64(&m.hits)
			var rate float64
			if g > 0 {
				rate = float64(h) / float64(g)
			}
			logger.WithFields(logrus.Fields{
				"layer":    "memory",
				"gets":     g,
				"hits":     h,
				"hit_rate": rate,
			}).Info("ncache 内存命中率")
		}
	}
}

// MGet 批量读取：逐键读取并判断过期，返回命中项与未命中 keys。
func (m *memoryLayer) MGet(ctx context.Context, keys []string) (map[string][]byte, []string, error) {
	hits := make(map[string][]byte, len(keys))
	var missing []string
	if len(keys) == 0 {
		return hits, missing, nil
	}
	now := time.Now()
	for _, key := range keys {
		atomic.AddUint64(&m.gets, 1)
		v, ok := m.m.Load(key)
		if !ok {
			missing = append(missing, key)
			continue
		}
		it := v.(memoryItem)
		if !it.expireAt.IsZero() && now.After(it.expireAt) {
			m.m.Delete(key)
			missing = append(missing, key)
			continue
		}
		hits[key] = it.v
		atomic.AddUint64(&m.hits, 1)
	}
	return hits, missing, nil
}

// MSet 批量写入：统一 TTL；写入后进行一次软上限清理。
func (m *memoryLayer) MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	for k, v := range items {
		m.m.Store(k, memoryItem{v: v, expireAt: expireAt})
	}
	// 软上限：超过 maxEntries 后触发一次过期清理（不保证严格约束）
	if m.max > 0 {
		count := 0
		m.m.Range(func(_, _ any) bool { count++; return count <= m.max+1 })
		if count > m.max {
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
