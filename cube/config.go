package cube

import (
	"fmt"
	"time"
)

var DefaultConfig = Config{
	BaseURL:    "",
	APIKey:     "",
	BucketName: "",

	Timeout: 10 * time.Second,

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
	BaseURL    string `mapstructure:"base_url"`    // 基础 URL
	APIKey     string `mapstructure:"api_key"`     // 应用密钥
	BucketName string `mapstructure:"bucket_name"` // 存储桶名称

	Timeout time.Duration `mapstructure:"timeout"` // HTTP 请求超时时间

	// HTTP Client Transport配置
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
}

type CubeError struct {
	Code int
	Msg  string
}

// Error 实现 error 接口
func (e *CubeError) Error() string {
	return fmt.Sprintf("Cube 业务错误 [Code: %d]: %s", e.Code, e.Msg)
}

// NewCubeError 是 CubeError 的工厂方法
func NewCubeError(code int, message string) *CubeError {
	return &CubeError{
		Code: code,
		Msg:  message,
	}
}
