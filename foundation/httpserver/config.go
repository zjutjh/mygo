package httpserver

import "time"

var DefaultConfig = Config{
	Addr: ":8888",

	ShutdownWaitTimeout: 10 * time.Second,

	Pprof: false,

	H2C: H2CConfig{
		Enable: true,
	},

	Log: LogConfig{
		AccessFilename: "./logs/access.log",
		ErrorFilename:  "./logs/error.log",
		MaxSize:        100,
		MaxAge:         7,
		MaxBackups:     14,
		LocalTime:      false,
		Compress:       false,
	},
	// Gin: GinConfig{},
}

type Config struct {
	Addr string `mapstructure:"addr"`

	ShutdownWaitTimeout time.Duration `mapstructure:"shutdown_wait_timeout"`

	Pprof bool `mapstructure:"pprof"`

	H2C H2CConfig `mapstructure:"h2c"`

	Log LogConfig `mapstructure:"log"`
	// Gin GinConfig `mapstructure:"gin"`
}

// H2CConfig 明文 HTTP/2 配置
type H2CConfig struct {
	Enable bool `mapstructure:"enable"`
}

type LogConfig struct {
	AccessFilename string `mapstructure:"access_filename"` // AccessFilename 日志文件路径
	ErrorFilename  string `mapstructure:"error_filename"`  // ErrorFilename 日志文件路径
	MaxSize        int    `mapstructure:"max_size"`        // MaxSize 触发日志切割大小 单位 MB
	MaxAge         int    `mapstructure:"max_age"`         // MaxAge 日志切割后文件保留天数
	MaxBackups     int    `mapstructure:"max_backups"`     // MaxBackups 日志切割后文件保留数量
	LocalTime      bool   `mapstructure:"local_time"`      // LocalTime 日志切割文件是否采用服务器本地时间
	Compress       bool   `mapstructure:"compress"`        // Compress 日志切割后是否对归档文件进行压缩
}

// Gin配置 有需要时再补充
type GinConfig struct {
	// RedirectTrailingSlash bool `mapstructure:"redirect_trailing_slash"`
	// ......
}
