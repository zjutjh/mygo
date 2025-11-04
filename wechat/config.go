package wechat

var DefaultConfig = Config{
	Redis: "",
	Resty: "",

	AppID:         "",
	AppSecret:     "",
	RedirectURL:   "",
	AESSecret:     "",
	TokenCacheKey: "wechat_access_token",
	MessageSuffix: "",
}

type Config struct {
	Redis string `mapstructure:"redis" json:"redis" yaml:"redis"`
	Resty string `mapstructure:"resty" json:"resty" yaml:"resty"`

	Enable        bool   `mapstructure:"enable"`
	AppID         string `mapstructure:"appid"`
	AppSecret     string `mapstructure:"secret"`
	RedirectURL   string `mapstructure:"redirect_url"`
	AESSecret     string `mapstructure:"aes_secret"`
	TokenCacheKey string `mapstructure:"token_cache_key"`
	MessageSuffix string `mapstructure:"message_suffix" json:"message_suffix" yaml:"message_suffix"`
}
