package officialAccount

import (
	"time"

	"github.com/ArtisanCloud/PowerLibs/v3/cache"
)

const (
	DriverRedis  = "redis"
	DriverMemory = "memory"
)

var DefaultConfig = Config{
	AppID:     "",
	Secret:    "",
	Token:     "",
	AESKey:    "",
	HttpDebug: false,
	Debug:     false,
	Log: LogConfig{
		Level:  "info",
		File:   "./logs/app.log",
		Error:  "./logs/app.log",
		Stdout: false,
	},
	Driver: DriverMemory,
	Redis:  "",
	MemCache: MemCacheConfig{
		Prefix:        "",
		DefaultExpire: cache.DEFAULT_EXPIRES_IN,
		Namespace:     "",
	},
}

type LogConfig struct {
	Level  string `mapstructure:"level" json:"level" yaml:"level"`
	File   string `mapstructure:"file" json:"file" yaml:"file"`
	Error  string `mapstructure:"error" json:"error" yaml:"error"`
	Stdout bool   `mapstructure:"stdout" json:"stdout" yaml:"stdout"`
}

type Config struct {
	AppID     string         `mapstructure:"appid" json:"appid" yaml:"appid"`
	Secret    string         `mapstructure:"secret" json:"secret" yaml:"secret"`
	Token     string         `mapstructure:"token" json:"token" yaml:"token"`
	AESKey    string         `mapstructure:"aes_key" json:"aes_key" yaml:"aes_key"`
	HttpDebug bool           `mapstructure:"http_debug" json:"http_debug" yaml:"http_debug"`
	Debug     bool           `mapstructure:"debug" json:"debug" yaml:"debug"`
	Log       LogConfig      `mapstructure:"log" json:"log" yaml:"log"`
	Driver    string         `mapstructure:"driver" json:"driver" yaml:"driver"`
	Redis     string         `mapstructure:"redis" json:"redis" yaml:"redis"`
	MemCache  MemCacheConfig `mapstructure:"mem_cache" json:"mem_cache" yaml:"mem_cache"`
}

type MemCacheConfig struct {
	Prefix        string        `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	DefaultExpire time.Duration `mapstructure:"default_expire" json:"default_expire" yaml:"default_expire"`
	Namespace     string        `mapstructure:"namespace" json:"namespace" yaml:"namespace"`
}
