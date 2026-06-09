package ncache

import "github.com/redis/go-redis/v9"

type options struct {
	redisClient redis.UniversalClient
}

type Option func(*options)

func defaultOptions() *options {
	return &options{}
}

// 如果引用了WithRedis选项，则启用Redis缓存
func WithRedis(cli redis.UniversalClient) Option {
	return func(o *options) {
		o.redisClient = cli
	}
}
