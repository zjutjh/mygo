package miniprogram

import (
	"time"
)

var DefaultConfig = Config{
	AppId:     "",
	AppSecret: "",
	Token:     "",
	AesKey:    "",

	HttpDebug: true,

	Log: LogConfig{
		Level:  "debug",
		File:   "",
		Error:  "",
		Stdout: false,
	},

	Redis:       "",
	MaxActive:   10,
	MaxIdle:     10,
	IdleTimeout: 60 * time.Second,
}

type Config struct {
	//小程序配置
	AppId     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	Token     string `mapstructure:"token"`
	AesKey    string `mapstructure:"aes_key"`

	HttpDebug bool `mapstructure:"http_debug"`

	Log LogConfig `mapstructure:"log"`
	//Redis配置
	Redis string `mapstructure:"redis"`

	MaxActive   int           `mapstructure:"max_active"`
	MaxIdle     int           `mapstructure:"max_idle"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	File   string `mapstructure:"file"`
	Error  string `mapstructure:"error"`
	Stdout bool   `mapstructure:"stdout"`
}
