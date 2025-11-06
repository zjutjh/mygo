package lock

import (
	"context"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	mutex *redsync.Mutex
}

func New(client redis.UniversalClient, conf Config) *RedisLock {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(conf.RedisKey,
		redsync.WithExpiry(conf.Expiry),
		redsync.WithRetryDelay(conf.RetryDelay),
		redsync.WithTries(conf.RetryCount+1))
	return &RedisLock{mutex: mutex}
}

func (l *RedisLock) Lock(ctx context.Context) error {
	return l.mutex.LockContext(ctx)
}

func (l *RedisLock) Unlock(ctx context.Context) error {
	_, err := l.mutex.UnlockContext(ctx)
	return err
}

func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	err := l.mutex.TryLockContext(ctx)
	return err == nil, err
}
