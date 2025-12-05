package cube

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nesty"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CubeClient struct {
	conf   Config
	client *resty.Client
}

// UploadResponse 定义 Cube 文件上传响应体
type UploadResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ObjectKey string `json:"object_key"`
	} `json:"data"`
}

// DeleteResponse 定义 Cube 文件删除响应体
type DeleteResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// New 创建一个 CubeClient
func New(conf Config) *CubeClient {
	return &CubeClient{
		conf:   conf,
		client: nesty.Pick(conf.Resty),
	}
}

// UploadFile 上传文件到存储立方
func (c *CubeClient) UploadFile(ctx context.Context, filename string, reader io.Reader, location string, convertWebp, useUUID bool) (*UploadResponse, error) {
	bucketName := c.conf.BucketName
	if bucketName == "" {
		return nil, fmt.Errorf("%w: 未指定存储桶", kit.ErrRequestInvalidParamter)
	}

	form := map[string]string{
		"bucket":       bucketName,
		"location":     location,
		"convert_webp": strconv.FormatBool(convertWebp),
		"use_uuid":     strconv.FormatBool(useUUID),
	}

	var result UploadResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Key", c.conf.APIKey).
		SetFileReader("file", filename, reader).
		SetFormData(form).
		SetResult(&result).
		Post(strings.TrimRight(c.conf.BaseURL, "/") + "/api/upload")
	if err != nil {
		return nil, fmt.Errorf("上传文件请求错误: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("上传文件消息状态码错误: %v", resp.StatusCode())
	}
	return &result, nil
}

// DeleteFile 删除文件
func (c *CubeClient) DeleteFile(objectKey string) (*DeleteResponse, error) {
	bucketName := c.conf.BucketName
	if bucketName == "" {
		return nil, fmt.Errorf("%w: 未指定存储桶", kit.ErrRequestInvalidParamter)
	}

	var result DeleteResponse
	resp, err := c.client.R().
		SetHeader("Key", c.conf.APIKey).
		SetQueryParams(map[string]string{
			"bucket":     bucketName,
			"object_key": objectKey,
		}).
		SetResult(&result).
		Delete(strings.TrimRight(c.conf.BaseURL, "/") + "/api/delete")
	if err != nil {
		return nil, fmt.Errorf("删除文件请求错误: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("删除文件消息状态码错误: %v", resp.StatusCode())
	}
	return &result, nil
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
