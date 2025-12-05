package cube

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/zjutjh/mygo/kit"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CubeClient struct {
	conf   Config
	client *resty.Client
}

// UploadResponse 定义 Cube 文件上传 API 的标准响应结构
type UploadResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ObjectKey string `json:"object_key"`
	} `json:"data"`
}

// New 创建一个 CubeClient。
func New(conf Config) *CubeClient {
	// 初始化HTTP Client实例
	hc := &http.Client{
		Timeout: conf.Timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   conf.DialContextTimeout,
				KeepAlive: conf.DialContextKeepAlive,
			}).DialContext,
			TLSHandshakeTimeout:    conf.TLSHandshakeTimeout,
			DisableKeepAlives:      conf.DisableKeepAlives,
			DisableCompression:     conf.DisableCompression,
			MaxIdleConns:           conf.MaxIdleConns,
			MaxIdleConnsPerHost:    conf.MaxIdleConnsPerHost,
			MaxConnsPerHost:        conf.MaxConnsPerHost,
			IdleConnTimeout:        conf.IdleConnTimeout,
			ResponseHeaderTimeout:  conf.ResponseHeaderTimeout,
			ExpectContinueTimeout:  conf.ExpectContinueTimeout,
			MaxResponseHeaderBytes: conf.MaxResponseHeaderBytes,
			WriteBufferSize:        conf.WriteBufferSize,
			ReadBufferSize:         conf.ReadBufferSize,
			ForceAttemptHTTP2:      conf.ForceAttemptHTTP2,
		},
	}

	client := resty.NewWithClient(hc).
		SetBaseURL(strings.TrimRight(conf.BaseURL, "/")).
		SetHeader("Key", conf.APIKey)

	return &CubeClient{conf: conf, client: client}
}

// UploadFile 上传文件到存储立方
func (c *CubeClient) UploadFile(filename string, reader io.Reader, location string, convertWebp, useUUID bool) (string, error) {
	bucketName := c.conf.BucketName
	if bucketName == "" {
		return "", fmt.Errorf("%w: 未指定存储桶", kit.ErrRequestInvalidParamter)
	}

	form := map[string]string{
		"bucket":       bucketName,
		"location":     location,
		"convert_webp": strconv.FormatBool(convertWebp),
		"use_uuid":     strconv.FormatBool(useUUID),
	}

	var result UploadResponse
	resp, err := c.client.R().
		SetFileReader("file", filename, reader).
		SetFormData(form).
		SetResult(&result).
		Post("/api/upload")
	if err != nil {
		return "", fmt.Errorf("上传文件请求错误: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("上传文件消息状态码错误: %w", err)
	}
	if result.Code != 200 {
		return "", NewCubeError(result.Code, result.Msg)
	}
	return result.Data.ObjectKey, nil
}

// DeleteFile 删除文件
func (c *CubeClient) DeleteFile(objectKey string) error {
	bucketName := c.conf.BucketName
	if bucketName == "" {
		return fmt.Errorf("%w: 未指定存储桶", kit.ErrRequestInvalidParamter)
	}

	var result UploadResponse
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"bucket":     bucketName,
			"object_key": objectKey,
		}).
		SetResult(&result).
		Delete("/api/delete")
	if err != nil {
		return fmt.Errorf("删除文件请求错误: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("删除文件消息状态码错误: %w", err)
	}
	if result.Code != 200 {
		return NewCubeError(result.Code, result.Msg)
	}
	return nil
}

// GetFileURL 返回文件访问 URL
func (c *CubeClient) GetFileURL(objectKey string, thumbnail bool) string {
	baseURL := strings.TrimRight(c.conf.BaseURL, "/")
	params := url.Values{}
	params.Add("bucket", c.conf.BucketName)
	params.Add("object_key", objectKey)
	if thumbnail {
		params.Add("thumbnail", "true")
	}
	return fmt.Sprintf("%s/api/file?%s", baseURL, params.Encode())
}
