package miniProgram

import (
	"time"
)

const (
	CacheDriverRedis  = "redis"
	CacheDriverMemory = "memory"
)

var DefaultConfig = Config{
	AppID:  "",
	Secret: "",

	AppKey:  "",
	OfferID: "",

	Token:             "",
	AESKey:            "",
	ComponentAppID:    "",
	ComponentAppToken: "",
	StableTokenMode:   false,
	RefreshToken:      "",

	ResponseType: "",
	Http: HttpConfig{
		Timeout:  0,
		BaseURI:  "https://api.weixin.qq.com/",
		ProxyURI: "",
	},
	Cache: CacheConfig{
		Driver: CacheDriverMemory,
		MemCache: MemCacheConfig{
			Namespace:       "ac.go.power",
			DefaultLifeTime: 1500,
			Prefix:          "",
		},
	},

	HttpDebug: false,
	Debug:     false,

	Log:   "",
	Resty: "",
	Redis: "",
}

type Config struct {
	AppID  string `mapstructure:"app_id"`
	Secret string `mapstructure:"secret"`

	AppKey  string `mapstructure:"app_key"`
	OfferID string `mapstructure:"offer_id"`

	Token             string `mapstructure:"token"`
	AESKey            string `mapstructure:"aes_key"`
	ComponentAppID    string `mapstructure:"component_app_id"`
	ComponentAppToken string `mapstructure:"component_app_token"`
	StableTokenMode   bool   `mapstructure:"stable_token_mode"`
	RefreshToken      string `mapstructure:"refresh_token"`

	ResponseType string      `mapstructure:"response_type"`
	Http         HttpConfig  `mapstructure:"http"`
	Cache        CacheConfig `mapstructure:"cache"`

	HttpDebug bool `mapstructure:"http_debug"`
	Debug     bool `mapstructure:"debug"`

	// 基础依赖组件实例配置
	Log   string `mapstructure:"log"`
	Resty string `mapstructure:"resty"`
	Redis string `mapstructure:"redis"`
}

type HttpConfig struct {
	Timeout  float64 `mapstructure:"timeout"`
	BaseURI  string  `mapstructure:"base_uri"`
	ProxyURI string  `mapstructure:"proxy_uri"`
}

type CacheConfig struct {
	Driver   string         `mapstructure:"driver"`
	MemCache MemCacheConfig `mapstructure:"mem_cache"`
}

type MemCacheConfig struct {
	Namespace       string        `mapstructure:"namespace"`
	DefaultLifeTime time.Duration `mapstructure:"default_life_time"`
	Prefix          string        `mapstructure:"prefix"`
}
