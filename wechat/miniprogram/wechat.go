package miniprogram

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/response"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nedis"
)

type WeChatService interface {
	GetAccessToken() (string, error)
	GetMiniProgram() *miniProgram.MiniProgram
	GetConfig() *Config
}
type WeChat struct {
	conf        Config
	miniProgram *miniProgram.MiniProgram
	client      resty.Client
}

// New 创建微信服务实例
func New(conf Config) *WeChat {
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

	// 初始化resty Client
	client := resty.NewWithClient(hc)

	var kernelCache cache.CacheInterface
	// 配置缓存
	gr := cache.NewGRedis(&redis.UniversalOptions{})
	gr.Pool = nedis.Pick(conf.Redis)
	kernelCache = gr

	// 创建小程序实例
	mp, _ := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:        conf.AppId,
		Secret:       conf.AppSecret,
		ResponseType: response.TYPE_MAP,
		Token:        conf.Token,
		AESKey:       conf.AesKey,

		HttpDebug: conf.HttpDebug,
		Log: miniProgram.Log{
			Level:  conf.Log.Level,
			File:   conf.Log.File,
			Error:  conf.Log.Error,
			Stdout: conf.Log.Stdout,
		},

		// 应该使用AccessTokenCacheKey作为缓存键名

		Cache: kernelCache,
	})

	return &WeChat{
		conf:        conf,
		miniProgram: mp,
		client:      *client,
	}
}

type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// Code2Session 小程序登录身份验证
func (wc *WeChat) Code2Session(code string) (*Code2SessionResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("%w: code不能为空", kit.ErrRequestInvalidParamter)
	}

	resp, err := wc.client.R().SetQueryParams(map[string]string{
		"appid":      wc.conf.AppId,
		"secret":     wc.conf.AppSecret,
		"js_code":    code,
		"grant_type": "authorization_code",
	}).Get("https://api.weixin.qq.com/sns/jscode2session")

	if err != nil {
		return nil, fmt.Errorf("小程序登录请求错误: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("小程序登录状态码错误: %d", resp.StatusCode())
	}

	var result Code2SessionResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("小程序登录业务错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	if result.OpenID == "" {
		return nil, fmt.Errorf("获取OpenID失败: %s", string(resp.Body()))
	}

	return &result, nil
}

// GetAccessToken 获取AccessToken
func (wc *WeChat) GetAccessToken() (string, error) {

	// Redis缓存中获取access_token
	rdb := nedis.Pick(wc.conf.RedisInstance)
	if rdb != nil {
		if val, err := rdb.Get(context.Background(), wc.conf.AccessTokenCacheKey).Result(); err == nil {
			return val, nil
		}
	}

	// 重新获取access_token
	resp, err := wc.client.R().
		SetQueryParams(map[string]string{
			"grant_type": "client_credential",
			"appid":      wc.conf.AppId,
			"secret":     wc.conf.AppSecret,
		}).
		SetHeader("Accept", "application/json").
		Get("https://api.weixin.qq.com/cgi-bin/token")

	if err != nil {
		return "", fmt.Errorf("获取AccessToken请求错误: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("获取AccessToken状态码错误: %d", resp.StatusCode())
	}

	jsonData := string(resp.Body())
	accessToken := gjson.Get(jsonData, "access_token").String()
	if accessToken == "" {
		return "", fmt.Errorf("获取AccessToken失败: %s", jsonData)
	}

	// 缓存access_token
	expireTime := gjson.Get(jsonData, "expires_in").Int() - 60*30
	if rdb != nil {
		rdb.Set(context.Background(), wc.conf.AccessTokenCacheKey, accessToken, time.Duration(expireTime)*time.Second)
	}

	return accessToken, nil
}
func (wc *WeChat) GetMiniProgram() *miniProgram.MiniProgram {
	return wc.miniProgram
}
func (wc *WeChat) GetConfig() *Config {
	return &wc.conf
}

// DecryptUserData 添加用户敏感信息解密（手机号、用户详细信息等）
func (wc *WeChat) DecryptUserData(encryptedData, iv, sessionKey string) (map[string]interface{}, error) {
	aesKey, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("sessionKey解码失败: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("encryptedData解码失败: %w", err)
	}

	aesIV, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("iv解码失败: %w", err)
	}

	// AES-128-CBC解密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("创建AES密码器失败: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("密文长度不足")
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, aesIV)
	mode.CryptBlocks(ciphertext, ciphertext)

	// PKCS#7解填充
	paddingLen := int(ciphertext[len(ciphertext)-1])
	if paddingLen > len(ciphertext) || paddingLen > aes.BlockSize {
		return nil, fmt.Errorf("无效的填充长度")
	}

	decrypted := ciphertext[:len(ciphertext)-paddingLen]

	var userData map[string]interface{}
	if err := json.Unmarshal(decrypted, &userData); err != nil {
		return nil, fmt.Errorf("解析解密数据失败: %w", err)
	}
	// 验证appid
	if appid, ok := userData["watermark"].(map[string]interface{})["appid"].(string); !ok || appid != wc.conf.AppId {
		return nil, fmt.Errorf("无效的appid: %s", appid)
	}

	return userData, nil
}
