package lock

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

func New(client redis.UniversalClient, conf Config) *redsync.Mutex {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(conf.RedisKey,
		redsync.WithExpiry(conf.Expiry),
		redsync.WithRetryDelay(conf.RetryDelay),
		redsync.WithTries(conf.RetryCount+1))
	return mutex
}
