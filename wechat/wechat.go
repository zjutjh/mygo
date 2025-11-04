package wechat

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"

	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nedis"
)

type Wechat struct {
	conf   Config
	client *resty.Client
}

// New 以指定配置创建实例
func New(conf Config) *Wechat {
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

	return &Wechat{
		conf:   conf,
		client: client,
	}
}

func (w *Wechat) Oauth(redirectURL ...string) string {
	if !w.conf.Enable {
		return ""
	}

	targetURL := w.conf.RedirectURL
	if len(redirectURL) > 0 && redirectURL[0] != "" {
		targetURL = redirectURL[0]
	}

	return "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" +
		w.conf.AppID +
		"&redirect_uri=" + url.QueryEscape(targetURL) +
		"&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect"
}

func (w *Wechat) GetOpenID(code string) (string, error) {
	if !w.conf.Enable {
		return "", fmt.Errorf("%w: 微信功能未启用", kit.ErrNotFound)
	}

	if code == "" {
		return "", fmt.Errorf("%w: code不能为空", kit.ErrRequestInvalidParamter)
	}

	getOpenIDURL := "https://api.weixin.qq.com/sns/oauth2/access_token"
	resp, err := w.client.R().SetQueryParams(map[string]string{
		"appid":      w.conf.AppID,
		"secret":     w.conf.AppSecret,
		"code":       code,
		"grant_type": "authorization_code",
	}).Get(getOpenIDURL)

	if err != nil {
		return "", fmt.Errorf("获取OpenID请求错误: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("获取OpenID状态码错误: %d", resp.StatusCode())
	}

	jsonData := string(resp.Body())
	openid := gjson.Get(jsonData, "openid").String()
	if openid == "" {
		return "", fmt.Errorf("获取OpenID失败: %s", jsonData)
	}

	return openid, nil
}

func (w *Wechat) GetAccessToken() (string, error) {
	if !w.conf.Enable {
		return "", fmt.Errorf("%w: 微信功能未启用", kit.ErrNotFound)
	}

	// Redis缓存中获取access_token
	rdb := nedis.Pick(w.conf.TokenCacheKey)
	if rdb != nil {
		if val, err := rdb.Get(context.Background(), w.conf.TokenCacheKey).Result(); err == nil {
			return val, nil
		}
	}

	// 重新获取access_token
	resp, err := w.client.R().
		SetQueryParams(map[string]string{
			"grant_type": "client_credential",
			"appid":      w.conf.AppID,
			"secret":     w.conf.AppSecret,
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
		rdb.Set(context.Background(), w.conf.TokenCacheKey, accessToken, time.Duration(expireTime)*time.Second)
	}

	return accessToken, nil
}

func (w *Wechat) SendMessage(message, openID string) error {
	if !w.conf.Enable {
		return nil
	}

	if message == "" {
		return fmt.Errorf("%w: 消息内容不能为空", kit.ErrRequestInvalidParamter)
	}

	if openID == "" {
		return fmt.Errorf("%w: 接收者OpenID不能为空", kit.ErrRequestInvalidParamter)
	}

	accessToken, err := w.GetAccessToken()
	if err != nil {
		return fmt.Errorf("获取AccessToken失败: %w", err)
	}

	msgContent := message + "\n---\n因为微信的限制，请回复'收到'以确保后续消息的正常接收"

	params := map[string]interface{}{
		"touser":  openID,
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": msgContent,
		},
	}

	resp, err := w.client.R().
		SetBody(params).
		Post("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken)

	if err != nil {
		return fmt.Errorf("发送微信消息请求错误: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("发送微信消息状态码错误: %d", resp.StatusCode())
	}

	jsonData := string(resp.Body())
	errcode := gjson.Get(jsonData, "errcode").Int()
	if errcode != 0 {
		return fmt.Errorf("发送微信消息业务错误: %s", jsonData)
	}

	return nil
}