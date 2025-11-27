package cube

import (
	"fmt"
	"time"
)

// CubeConfig 用于配置客户端连接信息。
type CubeConfig struct {
	BaseURL           string        `json:"baseUrl" yaml:"baseUrl"`           // 基础 URL (对应 Java 的 baseUrl)
	APIKey            string        `json:"apiKey" yaml:"apiKey"`             // 应用密钥 (对应 Java 的 apiKey)
	DefaultBucketName string        `json:"bucketName" yaml:"bucketName"`     // 默认存储桶名称 (对应 Java 的 bucketName)
	TimeoutSeconds    int           `json:"timeout" yaml:"timeout"`           // 请求超时时间 (Go 惯例)
	Debug             bool          `json:"debug" yaml:"debug"`               // 调试模式 (Go 惯例)
}

type CubeError struct {
	Code int    // 业务错误码
	Msg  string // 业务错误信息
}

// Error 实现了 Go 的 error 接口，用于打印错误信息。
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

// UploadResponse 定义 Cube 文件上传 API 的标准响应结构。
type UploadResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ObjectKey string `json:"objectKey"`           // 小驼峰
		ObjectKey_ string `json:"object_key,omitempty"` // snake_case
		URL       string `json:"url"`
		FileID    string `json:"fileId"`             // 小驼峰
		FileID_   string `json:"file_id,omitempty"`   // snake_case
	} `json:"data"`
}

// GetObjectKey 提供统一访问接口
func (r *UploadResponse) GetObjectKey() string {
	if r.Data.ObjectKey != "" {
		return r.Data.ObjectKey
	}
	return r.Data.ObjectKey_
}
// GetFileID 提供统一访问接口
func (r *UploadResponse) GetFileID() string {
	if r.Data.FileID != "" {
		return r.Data.FileID
	}
	return r.Data.FileID_
}