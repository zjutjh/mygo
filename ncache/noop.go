package ncache

import (
	"context"
	"time"
)

// noopCache 表示禁用状态的缓存实现：
// - Get 始终返回 ErrNotFound
// - Set/Delete 为 no-op
// - Remember 直接调用 loader 并不做存储
// 这样可以在全局关闭缓存时不改调用代码逻辑

type noopCache struct{}

func (noopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrNotFound
}

func (noopCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return nil
}

func (noopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (noopCache) Remember(ctx context.Context, key string, loader LoaderFunc, opts ...Option) ([]byte, error) {
	v, _, err := loader(ctx)
	return v, err
}
