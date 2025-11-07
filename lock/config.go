package lock

import "time"

var DefaultConfig = Config{
	DefaultLock: LockConfig{
		Expiry:     30 * time.Second,
		RetryDelay: 100 * time.Millisecond,
		RetryCount: 3,
	},
	Locks: make(map[string]LockConfig),
}

type Config struct {
	DefaultLock LockConfig            `mapstructure:"default"`
	Locks       map[string]LockConfig `mapstructure:"locks"`
}

type LockConfig struct {
	Expiry     time.Duration `mapstructure:"expiry"`
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	RetryCount int           `mapstructure:"retry_count"`
}
