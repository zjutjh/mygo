package ncache

import (
	"context"
	"time"
)

// Cache 抽象：提供基础的 Get/Set/Delete，以及带回源策略的 Remember。
type Cache interface {
	// Get 仅从缓存读取；未命中返回 ErrNotFound。
	Get(ctx context.Context, key string) ([]byte, error)
	// Set 写入缓存链（所有层），使用 ttl；返回第一个错误。
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	// Delete 删除所有层的 key。
	Delete(ctx context.Context, key string) error
	// Remember 缓存读取并在未命中时回源；根据策略防止击穿/穿透。
	Remember(ctx context.Context, key string, loader LoaderFunc, opts ...Option) ([]byte, error)
}

// LoaderFunc 回源函数：返回值、建议TTL（0表示使用默认TTL）、错误。
type LoaderFunc func(ctx context.Context) (val []byte, ttl time.Duration, err error)

// Option 运行时选项，覆盖配置。
type Option func(*rememberOpts)

// 回源策略（读取路径）。
type SourcePolicy string

const (
	PolicyCacheAside             SourcePolicy = "cache_aside"    // 经典旁路缓存，未命中调用 loader 并回填
	PolicyCacheAsideSingleFlight SourcePolicy = "cache_aside_sf" // 旁路缓存 + 单飞防击穿
)

// ErrNotFound 与 kit.ErrNotFound 对齐，由上层识别为未命中。
var ErrNotFound = errorNotFound{}

type errorNotFound struct{}

func (e errorNotFound) Error() string { return "资源不存在" }

// Layer 多级缓存层接口（L1/L2...），面向字节以便统一编码策略。
type Layer interface {
	Get(ctx context.Context, key string) (val []byte, ttlLeft time.Duration, ok bool, err error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// BatchLoaderFunc 批量回源：对找到的 key 返回对应的值与 TTL。
// 未找到的 key 不需要出现在返回 map 中（缺席即未命中）。
// 批量相关接口与类型已迁移到 batch.go，保持 base.go 精简。
