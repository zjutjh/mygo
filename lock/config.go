package lock

import "time"

var DefaultConfig = Config{
	RedisKey:   "",
	Expiry:     30 * time.Second,
	RetryDelay: 100 * time.Millisecond,
	RetryCount: 3,
}

type Config struct {
	RedisKey   string        `mapstructure:"redis_key"`
	Expiry     time.Duration `mapstructure:"expiry"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	RetryCount int           `mapstructure:"retry_count"`
}
