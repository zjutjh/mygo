package nmdb

import (
	"time"
)

type LogLevel int

const (
	Silent LogLevel = iota
	Error
	Warn
	Info
)

// MongoDB 配置
var DefaultConfig = Config{
	Host:     []string{"localhost:27017"},
	Username: "",
	Password: "",
	Database: "",
	AuthDB:   "admin",

	// 连接配置
	ConnectTimeout:         10 * time.Second,
	SocketTimeout:          10 * time.Second,
	ServerSelectionTimeout: 30 * time.Second,

	// 连接池配置
	MaxPoolSize:   100,
	MinPoolSize:   0,
	MaxConnecting: 2,
	MaxIdleTime:   10 * time.Minute,

	// 其他配置
	Direct:      false,
	RetryWrites: true,
	RetryReads:  true,

	// 读写关注配置
	ReadConcern:    "",
	WriteConcern:   "",
	ReadPreference: "",

	// 压缩配置
	ZstdLevel: 0,

	// 心跳检测
	HeartbeatInterval: 10 * time.Second,
	//日志配置
	OpenLogger:                true,
	Log:                       "",
	SlowThreshold:             200 * time.Millisecond,
	Colorful:                  false,
	IgnoreRecordNotFoundError: true,
	LogLevel:                  Warn,
}

type Config struct {
	Host     []string `mapstructure:"host"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	Database string   `mapstructure:"database"`
	AuthDB   string   `mapstructure:"auth_db"`

	// 连接配置
	ConnectTimeout         time.Duration `mapstructure:"connect_timeout"`
	SocketTimeout          time.Duration `mapstructure:"socket_timeout"`
	ServerSelectionTimeout time.Duration `mapstructure:"server_selection_timeout"`

	// 连接池配置
	MaxPoolSize   uint64        `mapstructure:"max_pool_size"`
	MinPoolSize   uint64        `mapstructure:"min_pool_size"`
	MaxConnecting uint64        `mapstructure:"max_connecting"`
	MaxIdleTime   time.Duration `mapstructure:"max_idle_time"`
	//其他配置
	Direct      bool     `mapstructure:"direct"`
	RetryWrites bool     `mapstructure:"retry_writes"`
	RetryReads  bool     `mapstructure:"retry_reads"`
	ReplicaSet  string   `mapstructure:"replica_set"`
	AppName     string   `mapstructure:"app_name"`
	Compressors []string `mapstructure:"compressors"`

	// 读写关注配置
	ReadConcern    string `mapstructure:"read_concern"`
	WriteConcern   string `mapstructure:"write_concern"`
	ReadPreference string `mapstructure:"read_preference"`

	// 压缩配置
	ZstdLevel int `mapstructure:"zstd_level"`

	// 心跳检测
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`

	//日志配置
	OpenLogger                bool          `mapstructure:"open_logger"`
	Log                       string        `mapstructure:"log"`
	SlowThreshold             time.Duration `mapstructure:"slow_threshold"`
	Colorful                  bool          `mapstructure:"colorful"`
	IgnoreRecordNotFoundError bool          `mapstructure:"ignore_record_not_found_error"`
	LogLevel                  LogLevel      `mapstructure:"log_level"`
}
