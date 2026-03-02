package lock

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"

	"github.com/zjutjh/mygo/nedis"
)

// New 以指定配置创建实例
func New(conf Config) *redsync.Redsync {
	rdb := nedis.Pick(conf.Redis)
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	return rs
}
