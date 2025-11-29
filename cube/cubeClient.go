package cube

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// CubeClient 核心结构体。
// 封装了配置和 HTTP 客户端实例。
type CubeClient struct {
	Config *CubeConfig
	Resty  *resty.Client
}

// NewClient 是客户端的工厂函数
func NewClient(cfg *CubeConfig) *CubeClient {

	// 处理 nil 配置
	if cfg == nil {
		cfg = &CubeConfig{TimeoutSeconds: 10} // 使用 Int 秒数作为默认值
	}
	// 确保 BaseURL 已配置
	if cfg.BaseURL == "" {
		panic("cube: BaseURL 未配置")
	}

	// 确保 TimeoutSeconds 不为零
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = 10
	}

	// 将秒数转换为 time.Duration 供 SetTimeout 使用
	timeoutDuration := time.Duration(cfg.TimeoutSeconds) * time.Second

	restyClient := resty.New()

	// 应用基础配置
	restyClient.SetBaseURL(cfg.BaseURL).
		SetTimeout(timeoutDuration). // 使用转换后的 Duration
		SetHeader("Key", cfg.APIKey)

	client := &CubeClient{
		Config: cfg,
		Resty:  restyClient,
	}

	return client
}

// UploadFile 上传文件到存储立方
func (c *CubeClient) UploadFile(
	localPath string,
	location string,
	convertWebp bool,
	useUuid bool) (string, error) {

	bucketName := c.Config.DefaultBucketName
	if bucketName == "" {
		return "", errors.New("默认存储桶名称未配置，请检查 CubeConfig")
	}

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("无法开启本地文档: %w", err)
	}
	defer file.Close()

	fileName := filepath.Base(localPath)

	// 构建 Multipart/Form-Data 请求
	var result UploadResponse
	req := c.Resty.R().
		SetFileReader("file", fileName, file). // 文件字段 (SetFileReader 自动处理 IO)
		SetFormData(map[string]string{
			"bucket":       bucketName,
			"convert_webp": fmt.Sprintf("%v", convertWebp), // boolean 转换为 string
			"use_uuid":     fmt.Sprintf("%v", useUuid),     // boolean 转换为 string
		}).
		SetResult(&result)
	if location != "" {
		req.SetFormData(map[string]string{"location": location})
	}

	// 发送 POST 请求
	resp, err := req.Post("/api/upload")

	if err != nil {
		return "", fmt.Errorf("cube 客户端请求失败: %w", err)
	}

	// 错误捕获

	// 检查是否有 HTTP 错误
	if resp.IsError() {
		return "", fmt.Errorf("cube HTTP 错误: %s (Status: %d). Response Body: %s",
			resp.Status(), resp.StatusCode(), resp.String()) // 打印响应体，帮助调试
	}

	// 检查是否有 JSON 解析错误（仅当 HTTP 状态码为成功时）
	if resp.IsSuccess() && resp.Error() != nil {
		// resp.Error() 返回一个错误对象，指示 SetResult() 失败
		return "", fmt.Errorf("cube 响应体JSON解析失败: %v. Response Body: %s",
			resp.Error(), resp.String())
	}

	if resp.IsError() {
		return "", fmt.Errorf("cube HTTP 错误: %s (Status: %d)", resp.Status(), resp.StatusCode())
	}

	// 检查业务错误
	if result.Code != 200 {
		return "", NewCubeError(result.Code, result.Msg)
	}

	objKey := result.Data.ObjectKey
	if objKey == "" {
		return "", NewCubeError(200500, "上传成功但 ObjectKey 缺失")
	}
	return objKey, nil
}

// DeleteFile 删除文件
func (c *CubeClient) DeleteFile(objectKey string) error {
	bucketName := c.Config.DefaultBucketName
	if bucketName == "" {
		return errors.New("默认存储桶名称未配置，请检查 CubeConfig")
	}

	// 构建 DELETE 请求，并将参数放入 Query Params
	var result UploadResponse
	resp, err := c.Resty.R().
		SetQueryParams(map[string]string{
			"bucket":     bucketName,
			"object_key": objectKey,
		}).
		SetResult(&result).
		Delete("/api/delete")
	if err != nil {
		return fmt.Errorf("cube 客户端删除请求失败: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("cube HTTP 错误: %s (Status: %d)", resp.Status(), resp.StatusCode())
	}

	// 检查业务 code
	if result.Code != 200 {
		return NewCubeError(result.Code, result.Msg)
	}

	return nil
}

// GetFileUrl 拼接获取文件的 URL
func (c *CubeClient) GetFileUrl(objectKey string, thumbnail bool) string {
	// 使用 Go 标准库 net/url 的 QueryEscape 对 objectKey 进行 URL 编码
	encodedKey := url.QueryEscape(objectKey)
	baseUrl := strings.TrimRight(c.Config.BaseURL, "/")
	finalUrl := fmt.Sprintf("%s/api/file?bucket=%s&object_key=%s",
		baseUrl,
		c.Config.DefaultBucketName,
		encodedKey)

	if thumbnail {
		finalUrl += "&thumbnail=true"
	}

	return finalUrl
}