package cube

import (
	"fmt"
)

// CubeConfig 用于配置客户端连接信息。
type CubeConfig struct {
	BaseURL           string        `json:"baseUrl" yaml:"baseUrl"`           // 基础 URL
	APIKey            string        `json:"apiKey" yaml:"apiKey"`             // 应用密钥 
	DefaultBucketName string        `json:"bucketName" yaml:"bucketName"`     // 默认存储桶名称 
	TimeoutSeconds    int           `json:"timeout" yaml:"timeout"`           // 请求超时时间 
}

type CubeError struct {
	Code int    // 业务错误码
	Msg  string // 业务错误信息
}

// Error 实现了 Go 的 error 接口，用于打印错误信息
func (e *CubeError) Error() string {
	return fmt.Sprintf("Cube API 业务错误 [Code: %d]: %s", e.Code, e.Msg)
}

// NewCubeError 带参错误构造函数 
func NewCubeError(code int, message string) *CubeError {
	return &CubeError{
		Code: code,
		Msg:  message,
	}
}

// UploadResponse 定义 Cube 文件上传 API 的标准响应结构
type UploadResponse struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data struct {
        ObjectKey string `json:"object_key"` 
        URL       string `json:"url"`
        FileID    string `json:"file_id"`   
    } `json:"data"`
}