package redis

import (
	"github.com/gin-contrib/sessions"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/session/redistore"
)

type Store interface {
	sessions.Store
}

func NewStore(client redis.UniversalClient, keyPairs ...[]byte) (Store, error) {
	s, err := redistore.NewRediStore(client, keyPairs...)
	if err != nil {
		return nil, err
	}
	return &store{s}, nil
}

type store struct {
	*redistore.RediStore
}

func (c *store) Options(options sessions.Options) {
	c.RediStore.Options = options.ToGorillaOptions()
}
