package miniprogram

import "time"

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

	Redis: RedisConfig{
		Size:    0,
		Network: "tcp",
		Address: "127.0.0.1:6379",
		DB:      "0",
	},
	EnableRedis: false,

	MemCache: MemCacheConfig{
		Prefix:        "",
		DefaultExpire: time.Duration(0),
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

	//Redis配置
	Redis       RedisConfig `mapstructure:"redis"`
	EnableRedis bool        `mapstructure:"enable_redis"`

	MemCache MemCacheConfig `mapstructure:"mem_cache"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	File   string `mapstructure:"file"`
	Error  string `mapstructure:"error"`
	Stdout bool   `mapstructure:"stdout"`
}

type RedisConfig struct {
	Size     int    `mapstructure:"size"`
	Network  string `mapstructure:"network"`
	Address  string `mapstructure:"address"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DB       string `mapstructure:"db"`
}

type MemCacheConfig struct {
	Prefix        string        `mapstructure:"prefix"`
	DefaultExpire time.Duration `mapstructure:"default_expire"`
	Namespace     string        `mapstructure:"namespace"`
}
