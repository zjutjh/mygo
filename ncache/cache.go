package ncache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/zjutjh/mygo/nedis"
)

// 多级缓存架构实现
type multiCache struct {
	conf   Config
	layers []Layer
	sf     singleflight.Group
}

// New 以配置创建 Cache 实例
func New(conf Config) (Cache, error) {
	// 如果未启用缓存，返回 no-op 实现，方便直接绕过缓存调试
	if !conf.Enable {
		// 返回一个 no-op cache，读总是未命中，写/删为 no-op
		return noopCache{}, nil
	}

	// 构建层列表
	var layers []Layer
	for _, lc := range conf.Layers {
		switch strings.ToLower(lc.Type) {
		case "memory":
			layers = append(layers, newMemoryLayer(lc.Memory))
		case "redis":
			scope := lc.Redis.Scope
			if scope == "" {
				scope = "redis"
			}
			// 确保对应的 Redis scope 已经由 nedis.Boot 提供，否则给出明确错误提示
			if !nedis.Exist(scope) {
				return nil, fmt.Errorf("ncache: 需要的 Redis scope[%s] 尚未初始化，请在启动阶段先调用 nedis.Boot(%q)", scope, scope)
			}
			layers = append(layers, newRedisLayer(scope))
		}
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("并没有配置任何缓存层")
	}
	return &multiCache{conf: conf, layers: layers}, nil
}

// Close 关闭所有层
func (c *multiCache) Close() error {
	for _, l := range c.layers {
		if cl, ok := l.(interface{ Close() error }); ok {
			_ = cl.Close()
		}
	}
	return nil
}

func (c *multiCache) namespaced(key string) string {
	if c.conf.Prefix == "" {
		return key
	}
	return c.conf.Prefix + key
}

// record 是统一编码结构，承载负缓存标记
type record struct {
	V []byte `json:"v,omitempty"`
	N bool   `json:"n"` // negative flag
}

func encodeRecord(v []byte, negative bool) []byte {
	b, _ := json.Marshal(record{V: v, N: negative})
	return b
}

func decodeRecord(b []byte) (rec record, err error) {
	err = json.Unmarshal(b, &rec)
	return
}

// Get 仅从缓存读取
func (c *multiCache) Get(ctx context.Context, key string) ([]byte, error) {
	nk := c.namespaced(key)
	// 自上而下查找
	for i, l := range c.layers {
		b, _, ok, err := l.Get(ctx, nk)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		rec, err := decodeRecord(b)
		if err != nil {
			// 数据损坏，认为未命中
			continue
		}
		if rec.N {
			return nil, ErrNotFound
		}
		// 如果第 i 层命中，下标 < i 的层会被调用 Set 回到上层
		for j := 0; j < i; j++ {
			_ = c.layers[j].Set(ctx, nk, b, c.conf.DefaultTTL)
		}
		return rec.V, nil
	}
	return nil, ErrNotFound
}

// Set 写入所有层
func (c *multiCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	nk := c.namespaced(key)
	if ttl <= 0 {
		ttl = c.conf.DefaultTTL
	}
	ttl = c.applyJitter(ttl)
	payload := encodeRecord(val, false)
	var firstErr error
	for _, l := range c.layers {
		if err := l.Set(ctx, nk, payload, ttl); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Delete 删除所有层
func (c *multiCache) Delete(ctx context.Context, key string) error {
	nk := c.namespaced(key)
	var firstErr error
	for _, l := range c.layers {
		if err := l.Delete(ctx, nk); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type rememberOpts struct {
	ttl    time.Duration
	policy SourcePolicy
}

// WithTTL 指定 Remember 写入的TTL
func WithTTL(ttl time.Duration) Option { return func(o *rememberOpts) { o.ttl = ttl } }

// WithPolicy 指定回源策略
func WithPolicy(p SourcePolicy) Option { return func(o *rememberOpts) { o.policy = p } }

// Remember 读取并在未命中时回源
func (c *multiCache) Remember(ctx context.Context, key string, loader LoaderFunc, opts ...Option) ([]byte, error) {
	// 先读
	if v, err := c.Get(ctx, key); err == nil {
		return v, nil
	}

	ro := rememberOpts{ttl: c.conf.DefaultTTL, policy: SourcePolicy(c.conf.Policy)}
	for _, opt := range opts {
		opt(&ro)
	}
	if ro.ttl <= 0 {
		ro.ttl = c.conf.DefaultTTL
	}

	switch ro.policy {
	case PolicyCacheAsideSingleFlight:
		// 旁路缓存 + 单飞
		nk := c.namespaced(key)
		v, err, _ := c.sf.Do(nk, func() (any, error) {
			// 执行双重检查
			if vv, err := c.Get(ctx, key); err == nil {
				return vv, nil
			}
			val, ttl, err := loader(ctx)
			if err != nil {
				if isNotFound(err) {
					// 负缓存
					_ = c.setNegative(ctx, key)
					return nil, ErrNotFound
				}
				return nil, err
			}
			if ttl <= 0 {
				ttl = ro.ttl
			}
			_ = c.Set(ctx, key, val, ttl)
			return val, nil
		})
		if err != nil {
			return nil, err
		}
		return v.([]byte), nil
	case PolicyCacheAside:
		// 简单的旁路缓存
		val, ttl, err := loader(ctx)
		if err != nil {
			if isNotFound(err) {
				_ = c.setNegative(ctx, key)
				return nil, ErrNotFound
			}
			return nil, err
		}
		if ttl <= 0 {
			ttl = ro.ttl
		}
		_ = c.Set(ctx, key, val, ttl)
		return val, nil
	default:
		// 默认执行单飞策略
		return c.Remember(ctx, key, loader, WithTTL(ro.ttl), WithPolicy(PolicyCacheAsideSingleFlight))
	}
}

// setNegative 写入负缓存
func (c *multiCache) setNegative(ctx context.Context, key string) error {
	nk := c.namespaced(key)
	// 负缓存TTL
	ttl := c.conf.NegativeTTL
	if ttl <= 0 {
		// 默认30秒
		ttl = 30 * time.Second
	}
	ttl = c.applyJitter(ttl)
	payload := encodeRecord(nil, true)
	var firstErr error
	for _, l := range c.layers {
		if err := l.Set(ctx, nk, payload, ttl); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// 应用 TTL 抖动
// 避免缓存雪崩
// 抖动范围由 JitterPercent 决定，取值范围 [-1.0, +1.0]
func (c *multiCache) applyJitter(ttl time.Duration) time.Duration {
	if c.conf.JitterPercent <= 0 {
		return ttl
	}
	// 抖动范围 [-p, +p]
	p := c.conf.JitterPercent
	delta := (rand.Float64()*2 - 1) * p
	adj := float64(ttl) * (1 + delta)
	if adj < float64(time.Second) {
		adj = float64(time.Second)
	}
	return time.Duration(adj)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	// 与本地 ErrNotFound 或 kit.ErrNotFound 文案兼容（尽量宽松）
	if err == ErrNotFound {
		return true
	}
	// 与其他系统的 ErrNotFound 文案兼容（尽量宽松）
	if strings.Contains(err.Error(), "资源不存在") {
		return true
	}
	return false
}

// 批量扩展：可选的层批量接口
// 批量实现已移动到 cache_batch.go，保持本文件聚焦单键逻辑。
