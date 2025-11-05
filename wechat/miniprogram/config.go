package miniprogram

import (
	"time"
)

var DefaultConfig = Config{
	AppId:               "",
	AppSecret:           "",
	Token:               "",
	AesKey:              "",
	RedisInstance:       "",
	AccessTokenCacheKey: "",

	Timeout: 5 * time.Second,

	HttpDebug: true,

	Log: LogConfig{
		Level:  "debug",
		File:   "/Users/User/wechat/mini-program/info.log",
		Error:  "/Users/User/wechat/mini-program/error.log",
		Stdout: false,
	},

	Redis:       "",
	MaxActive:   10,
	MaxIdle:     10,
	IdleTimeout: 60 * time.Second,

	TLSHandshakeTimeout:    0,
	DisableKeepAlives:      false,
	DisableCompression:     false,
	MaxIdleConns:           0,
	MaxIdleConnsPerHost:    200,
	MaxConnsPerHost:        500,
	IdleConnTimeout:        30 * time.Second,
	ResponseHeaderTimeout:  0,
	ExpectContinueTimeout:  1 * time.Second,
	MaxResponseHeaderBytes: 0,
	WriteBufferSize:        0,
	ReadBufferSize:         0,
	ForceAttemptHTTP2:      true,
	DialContextTimeout:     30 * time.Second,
	DialContextKeepAlive:   30 * time.Second,
}

type Config struct {
	//小程序配置
	AppId               string `mapstructure:"app_id"`
	AppSecret           string `mapstructure:"app_secret"`
	Token               string `mapstructure:"token"`
	AesKey              string `mapstructure:"aes_key"`
	RedisInstance       string `mapstructure:"redis_"`
	AccessTokenCacheKey string `mapstructure:"access_token_cache_key"`

	HttpDebug bool `mapstructure:"http_debug"`

	Log LogConfig `mapstructure:"log"`
	//Redis配置
	Redis string `mapstructure:"redis"`

	MaxActive   int           `mapstructure:"max_active"`
	MaxIdle     int           `mapstructure:"max_idle"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`

	// HTTP Client Transport
	TLSHandshakeTimeout    time.Duration `mapstructure:"tls_handshake_timeout"`
	DisableKeepAlives      bool          `mapstructure:"disable_keep_alives"`
	DisableCompression     bool          `mapstructure:"disable_compression"`
	MaxIdleConns           int           `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost    int           `mapstructure:"max_idle_conns_per_host"`
	MaxConnsPerHost        int           `mapstructure:"max_conns_per_host"`
	IdleConnTimeout        time.Duration `mapstructure:"idle_conn_timeout"`
	ResponseHeaderTimeout  time.Duration `mapstructure:"response_header_timeout"`
	ExpectContinueTimeout  time.Duration `mapstructure:"expect_continue_timeout"`
	MaxResponseHeaderBytes int64         `mapstructure:"max_response_header_bytes"`
	WriteBufferSize        int           `mapstructure:"write_buffer_size"`
	ReadBufferSize         int           `mapstructure:"read_buffer_size"`
	ForceAttemptHTTP2      bool          `mapstructure:"force_attempt_http2"`
	DialContextTimeout     time.Duration `mapstructure:"dial_context_timeout"`
	DialContextKeepAlive   time.Duration `mapstructure:"dial_context_keep_alive"`
	Timeout                time.Duration
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	File   string `mapstructure:"file"`
	Error  string `mapstructure:"error"`
	Stdout bool   `mapstructure:"stdout"`
}
