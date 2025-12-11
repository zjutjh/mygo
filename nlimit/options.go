package nlimit

import "github.com/redis/go-redis/v9"

type options struct {
	redisClient redis.UniversalClient
	keyPrefix   string
}

type Option func(*options)

func defaultOptions() *options {
	return &options{
		keyPrefix: "limiter",
	}
}

// WithRedis 配置使用 Redis 存储
func WithRedis(cli redis.UniversalClient) Option {
	return func(o *options) {
		o.redisClient = cli
	}
}

// WithPrefix 配置 Redis Key 前缀 (默认 "limiter")
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.keyPrefix = prefix
	}
}
