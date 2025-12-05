package cube

var DefaultConfig = Config{
	BaseURL:    "",
	APIKey:     "",
	BucketName: "",

	Resty: "",
}

type Config struct {
	BaseURL    string `mapstructure:"base_url"`    // 基础 URL
	APIKey     string `mapstructure:"api_key"`     // 应用密钥
	BucketName string `mapstructure:"bucket_name"` // 存储桶名称

	// 基础依赖组件实例配置
	Resty string `mapstructure:"resty"`
}
