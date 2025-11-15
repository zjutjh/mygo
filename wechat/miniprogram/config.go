package miniprogram

import "time"

const (
	DriverRedis  = "redis"
	DriverMemory = "memory"
)

var DefaultConfig = Config{
	AppId:     "",
	AppSecret: "",
	Token:     "",
	AesKey:    "",

	HttpDebug: true,
	Debug:     true,

	Log: LogConfig{
		Level:  "debug",
		File:   "./logs/app.log",
		Error:  "./logs/app.log",
		Stdout: false,
	},
	Driver: "memory",
	Redis:  "",
	MemCache: MemCacheConfig{
		Prefix:        "",
		DefaultExpire: 0,
		Namespace:     "",
	},
}

type Config struct {
	//小程序配置
	AppId     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	Token     string `mapstructure:"token"`
	AesKey    string `mapstructure:"aes_key"`

	HttpDebug bool `mapstructure:"http_debug"`
	Debug     bool `mapstructure:"debug"`

	Log LogConfig `mapstructure:"log"`
	//缓存配置
	Driver   string         `mapstructure:"driver"`
	Redis    string         `mapstructure:"redis"`
	MemCache MemCacheConfig `mapstructure:"mem_cache"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	File   string `mapstructure:"file"`
	Error  string `mapstructure:"error"`
	Stdout bool   `mapstructure:"stdout"`
}
type MemCacheConfig struct {
	Prefix        string        `mapstructure:"prefix"`
	DefaultExpire time.Duration `mapstructure:"default_expire"`
	Namespace     string        `mapstructure:"namespace"`
}
