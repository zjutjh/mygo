package ncache

import (
	"context"
	"time"
)

// BatchLoaderFunc 批量回源：对找到的 key 返回对应的值与 TTL。
// 未找到的 key 不需要出现在返回 map 中（缺席即未命中）。
type BatchLoaderFunc func(ctx context.Context, keys []string) (map[string]LoadedValue, error)

// LoadedValue 承载批量回源返回的值与 TTL（0 表示使用默认 TTL）。
type LoadedValue struct {
	Value []byte
	TTL   time.Duration
}

// BatchCache 为 ncache 的批量能力扩展接口，不破坏现有 Cache。
// 建议通过类型断言使用：
//
//	if bc, ok := cache.(ncache.BatchCache); ok { ... }
type BatchCache interface {
	Cache
	// 批量读取：返回命中项（key->value）与未命中的 keys。
	MGet(ctx context.Context, keys []string) (hits map[string][]byte, missing []string, err error)
	// 批量写入：统一 ttl（0 表示使用默认 TTL）。
	MSet(ctx context.Context, items map[string][]byte, ttl time.Duration) error
	// 批量记忆：对本次 missing keys 调用 loader 一次；回填命中与负缓存；返回最终命中的键值。
	MRemember(ctx context.Context, keys []string, loader BatchLoaderFunc, opts ...Option) (map[string][]byte, error)
}
