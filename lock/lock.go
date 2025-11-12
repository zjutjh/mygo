package lock

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/zjutjh/mygo/nedis"
)

func New(conf Config) *redsync.Redsync {
	redisClient := nedis.Pick(conf.Redis)
	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)
	return rs
}
