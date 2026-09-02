package oauth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"

	"github.com/zjutjh/mygo/nesty"
)

type OAuth struct {
	conf   Config
	client *resty.Client
}

// New 创建一个 OAuth 实例
func New(conf Config) *OAuth {
	return &OAuth{
		conf:   conf,
		client: nesty.Pick(conf.Resty),
	}
}

// Login 统一登陆
func (o *OAuth) Login(username, password string) ([]*http.Cookie, error) {
	// 0. 检查统一系统是否关闭
	if err := o.checkIsClosed(); err != nil {
		return nil, err
	}

	client := o.client.Clone()
	// 使用cookieJar管理cookie
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建Cookie Jar错误: %w", err)
	}
	client.SetCookieJar(cookieJar)

	// 1. 初始化请求
	resp, err := client.R().
		Get(o.conf.LoginURL)
	if err != nil {
		return nil, fmt.Errorf("初始化统一认证登录请求错误: %w", err)
	}

	resp, err = client.R().
		Get(o.conf.LoginURL)
	if err != nil {
		return nil, fmt.Errorf("获取统一认证登录页面错误: %w", err)
	}
	// 2. 登陆参数生成
	// 解析execution
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, fmt.Errorf("解析统一认证登录页面错误: %w", err)
	}
	execution := doc.
		Find("input[type=hidden][name=execution]").
		AttrOr("value", "")
	// 密码加密
	encPwd, err := o.getEncryptedPassword(client, password)
	if err != nil {
		return nil, fmt.Errorf("加密统一认证登录密码错误: %w", err)
	}

	loginParams := map[string]string{
		"username":  username,
		"password":  encPwd,
		"execution": execution,
		"_eventId":  "submit",
	}

	// 3. 发送登陆请求
	resp, err = client.R().
		SetFormData(loginParams).
		Post(o.conf.LoginURL)
	if err != nil {
		return nil, fmt.Errorf("发送统一认证登录请求错误: %w", err)
	}
	// 检查登陆信息
	if err = o.checkLogin(resp); err != nil {
		return nil, fmt.Errorf("统一认证登录失败: %w", err)
	}

	// 4. 提取指定域名下的session并构造cookie列表
	u, err := url.Parse(o.conf.PersonalCenterURL)
	if err != nil {
		return nil, fmt.Errorf("解析用户中心URL错误: %w", err)
	}
	return cookieJar.Cookies(u), nil
}

// GetUserInfo 登陆并获取用户信息
func (o *OAuth) GetUserInfo(username, password string) (cookies []*http.Cookie, userInfo UserInfo, err error) {
	cookies, err = o.Login(username, password)
	if err != nil {
		return cookies, userInfo, err
	}
	userData := struct {
		Data UserInfo `json:"data"`
	}{}
	_, err = o.client.R().
		SetCookies(cookies).
		SetResult(&userData).
		Get(o.conf.UserInfoURL)
	if err != nil {
		return cookies, userInfo, fmt.Errorf("获取统一认证用户信息错误: %w", err)
	}
	return cookies, userData.Data, nil
}
