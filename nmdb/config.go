package nmdb

import (
	"time"
)

var DefaultConfig = Config{
	Host:     "localhost",
	Port:     27017,
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
}

type Config struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	AuthDB   string `mapstructure:"auth_db"`

	// 连接配置
	ConnectTimeout         time.Duration `mapstructure:"connect_timeout"`
	SocketTimeout          time.Duration `mapstructure:"socket_timeout"`
	ServerSelectionTimeout time.Duration `mapstructure:"server_selection_timeout"`

	// 连接池配置
	MaxPoolSize   uint64        `mapstructure:"max_pool_size"`
	MinPoolSize   uint64        `mapstructure:"min_pool_size"`
	MaxConnecting uint64        `mapstructure:"max_connecting"`
	MaxIdleTime   time.Duration `mapstructure:"max_idle_time"`

	Direct      bool     `mapstructure:"direct"`
	RetryWrites bool     `mapstructure:"retry_writes"`
	RetryReads  bool     `mapstructure:"retry_reads"`
	ReplicaSet  string   `mapstructure:"replica_set"`
	AppName     string   `mapstructure:"app_name"`
	Compressors []string `mapstructure:"compressors"`
}
