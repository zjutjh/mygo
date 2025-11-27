package ncache

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	redis_store "github.com/eko/gocache/store/redis/v4"
	gocache "github.com/patrickmn/go-cache"
)

// 将 CacheManager 定义为接口，以便于扩展和替换实现
// 好用爱用
type CacheManager interface {
	Get(ctx context.Context, key any, returnObj interface{}) (interface{}, error)
	Set(ctx context.Context, key any, object interface{}, options ...store.Option) error
	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options ...store.InvalidateOption) error
	Clear(ctx context.Context) error
}

// cacheManager 实现 CacheManager 接口（组合了 Marshaler）
type cacheManager struct {
	*marshaler.Marshaler
	// 使用Marshaler来处理序列化和反序列化
}

// 创建一个新的 CacheManager 实例
func New(opts ...Option) CacheManager {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	// 初始化 Memory Store (L1 缓存)
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	memoryStore := gocache_store.NewGoCache(gocacheClient)
	memoryCache := cache.New[any](memoryStore)

	var cacheInstance cache.CacheInterface[any]

	if o.redisClient != nil {
		// 初始化 Redis Store (L2 缓存)
		redisStore := redis_store.NewRedis(o.redisClient)
		redisCache := cache.New[any](redisStore)

		// 使用 Chain 串联 Memory 和 Redis
		// 读取时：先查 Memory，没有则查 Redis，并回填 Memory
		// 写入/删除时：同时操作 Memory 和 Redis
		cacheInstance = cache.NewChain[any](memoryCache, redisCache)
	} else {
		// 仅使用 Memory
		cacheInstance = memoryCache
	}

	// 初始化 Marshaler
	marshal := marshaler.New(cacheInstance)

	return &cacheManager{Marshaler: marshal}
}
