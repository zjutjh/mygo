package wechat

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"

	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nesty"
)

var (
	rdb  redis.UniversalClient
	rctx context.Context
)

type Wechat struct {
	conf  Config
	resty *resty.Client
}

func New(conf Config) *Wechat {
	return &Wechat{
		conf:  conf,
		resty: nesty.Pick(conf.Resty),
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
	resp, err := w.resty.R().SetQueryParams(map[string]string{
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

	ctx := rctx
	if ctx == nil {
		ctx = context.Background()
	}
	if rdb != nil {
		if val, err := rdb.Get(ctx, w.conf.TokenCacheKey).Result(); err == nil {
			return val, nil
		}
	}

	// 重新获取access_token
	resp, err := w.resty.R().
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

	// 缓存access_token（提前30分钟过期）
	expireTime := gjson.Get(jsonData, "expires_in").Int() - 60*30
	if rdb != nil {
		_ = rdb.Set(ctx, w.conf.TokenCacheKey, accessToken, time.Duration(expireTime)*time.Second).Err()
	}

	return accessToken, nil
}

func (w *Wechat) SendMessageWithWechat(message, openID string) error {
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

	msgContent := message
	if tail := w.conf.MessageSuffix; tail != "" {
		msgContent = message + "\n---\n" + tail
	}

	params := map[string]interface{}{
		"touser":  openID,
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": msgContent,
		},
	}

	resp, err := w.resty.R().
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
