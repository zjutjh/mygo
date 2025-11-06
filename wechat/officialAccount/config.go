package officialAccount

// DefaultConfig 默认配置
var DefaultConfig = Config{
	AppID:     "",
	Secret:    "",
	Token:     "",
	AESKey:    "",
	HttpDebug: false,
	Debug:     false,
	Log: LogConfig{
		Level:  "info",
		File:   "",
		Error:  "",
		Stdout: false,
	},
}

type LogConfig struct {
	Level  string `mapstructure:"level" json:"level" yaml:"level"`
	File   string `mapstructure:"file" json:"file" yaml:"file"`
	Error  string `mapstructure:"error" json:"error" yaml:"error"`
	Stdout bool   `mapstructure:"stdout" json:"stdout" yaml:"stdout"`
}

type Config struct {
	AppID     string    `mapstructure:"appid" json:"appid" yaml:"appid"`
	Secret    string    `mapstructure:"secret" json:"secret" yaml:"secret"`
	Token     string    `mapstructure:"token" json:"token" yaml:"token"`
	AESKey    string    `mapstructure:"aes_key" json:"aes_key" yaml:"aes_key"`
	HttpDebug bool      `mapstructure:"http_debug" json:"http_debug" yaml:"http_debug"`
	Debug     bool      `mapstructure:"debug" json:"debug" yaml:"debug"`
	Log       LogConfig `mapstructure:"log" json:"log" yaml:"log"`
}
