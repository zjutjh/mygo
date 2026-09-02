package oauth

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// GetLoginMsg 获取登陆失败后页面上的提示语
func (o *OAuth) getLoginMsg(resp *resty.Response) string {
	re := regexp.MustCompile(`<span\s+id="msg">(.+?)</span>`)
	matches := re.FindStringSubmatch(resp.String())
	if len(matches) == 0 {
		return ""
	}
	// 删除span内部的标签
	re = regexp.MustCompile(`<[^>]*>`)
	msg := re.ReplaceAllString(matches[1], "")
	return msg
}

// CheckLogin 用于判断登陆是否成功
func (o *OAuth) checkLogin(resp *resty.Response) error {
	// 判断登陆是否成功
	destination := resp.RawResponse.Request.URL.String()
	if destination == o.conf.PersonalCenterURL {
		// 登陆成功后会跳转到用户中心
		return nil
	}

	// 判断是否需要修改密码
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return fmt.Errorf("解析统一认证登录响应错误: %w", err)
	}
	title := doc.Find("title").Text()
	if title == "修改密码" {
		return ErrEditPassword
	}

	// 判断失败原因
	msg := o.getLoginMsg(resp)
	switch msg {
	case o.conf.WrongPasswordMsg:
		return ErrWrongPassword
	case o.conf.WrongAccountMsg:
		return ErrWrongAccount
	case o.conf.NotActivatedMsg:
		return ErrNotActivated
	}
	return ErrOther
}

// CheckIsClosed 判断统一是否关闭
func (o *OAuth) checkIsClosed() error {
	resp, err := o.client.R().Get(o.conf.PersonalCenterURL)
	if err != nil {
		return fmt.Errorf("请求统一认证用户中心错误: %w", err)
	}

	// 校验httpStatusCode
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("统一认证用户中心状态码[%d]异常: %w", resp.StatusCode(), ErrClosed)
	}

	// fallback 校验内容
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return fmt.Errorf("解析统一认证用户中心响应错误: %w", err)
	}

	title := doc.Find("title").Text()
	if title == "Error 403.6" || title == "Error 403" {
		return ErrClosed
	}
	return nil
}
