package ncache

import "time"

// 默认配置
var DefaultConfig = Config{
	Enable:         true,
	Scope:          "cache",
	Prefix:         "",
	DefaultTTL:     5 * time.Minute,
	JitterPercent:  0.1,
	NegativeTTL:    30 * time.Second,
	StampedeWindow: 200 * time.Millisecond,
	Policy:         string(PolicyCacheAsideSingleFlight),
	Layers: []LayerConfig{
		{Type: "memory", Memory: MemoryConfig{MaxEntries: 10000, CleanInterval: time.Minute}},
		{Type: "redis", Redis: RedisConfig{Scope: "redis"}},
	},
}

type Config struct {
	// 组件开关与命名
	Enable bool   `mapstructure:"enable"`
	Scope  string `mapstructure:"scope"`
	// key 统一前缀
	Prefix string `mapstructure:"prefix"`
	// TTL 与抖动、负缓存 TTL、击穿保护窗口
	DefaultTTL     time.Duration `mapstructure:"default_ttl"`
	JitterPercent  float64       `mapstructure:"jitter_percent"`
	NegativeTTL    time.Duration `mapstructure:"negative_ttl"`
	StampedeWindow time.Duration `mapstructure:"stampede_window"`
	// 回源策略（读取）
	Policy string `mapstructure:"source_policy"`
	// 多级缓存层定义（自上而下：L1->Ln）
	Layers []LayerConfig `mapstructure:"layers"`
}

// 单层缓存配置
type LayerConfig struct {
	Type   string       `mapstructure:"type"`
	Memory MemoryConfig `mapstructure:"memory"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// 内存缓存层配置
type MemoryConfig struct {
	MaxEntries    int           `mapstructure:"max_entries"`
	CleanInterval time.Duration `mapstructure:"clean_interval"`
	// 命中率统计与日志输出（可选）。当 interval>0 且 log scope 存在时，会周期性输出命中率。
	StatsInterval time.Duration `mapstructure:"stats_interval"`
	StatsLog      string        `mapstructure:"stats_log"`
}

// Redis缓存层配置
type RedisConfig struct {
	// 复用 nedis 的 scope 名
	Scope string `mapstructure:"scope"`
}
