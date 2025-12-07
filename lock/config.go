package lock

var DefaultConfig = Config{
	Redis: "",
}

type Config struct {
	// 基础依赖组件实例配置
	Redis string `mapstructure:"redis"`
}
