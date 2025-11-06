package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	wechatOA "github.com/zjutjh/mygo/wechat/officialAccount"
)

func envBool(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "yes"
}

func main() {
	// 1) 读取配置（环境变量）
	conf := wechatOA.Config{
		AppID:     os.Getenv("WECHAT_APPID"),
		Secret:    os.Getenv("WECHAT_SECRET"),
		Token:     os.Getenv("WECHAT_TOKEN"),
		AESKey:    os.Getenv("WECHAT_AESKEY"),
		HttpDebug: envBool("WECHAT_HTTP_DEBUG", false),
		Debug:     envBool("WECHAT_DEBUG", false),
		Log: wechatOA.LogConfig{
			Level:  "info",
			File:   "",
			Error:  "",
			Stdout: true,
		},
	}

	// 2) 初始化客户端
	cli, err := wechatOA.New(conf)
	if err != nil {
		log.Fatalf("init official account error: %v", err)
	}

	// 3) 构造网页授权链接（本地演示）
	redirect := os.Getenv("WECHAT_REDIRECT_URI")
	if redirect == "" {
		redirect = "http://localhost:8080/callback"
	}
	// 保障是合法 URL
	if _, err := url.ParseRequestURI(redirect); err != nil {
		log.Fatalf("invalid WECHAT_REDIRECT_URI: %v", err)
	}

	// 根据官方文档拼接网页授权链接（前端通常直接拼接并跳转）
	// 文档：https://powerwechat.artisan-cloud.com/zh/official-account/oauth.html
	authURL := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		conf.AppID,
		url.QueryEscape(redirect),
		"snsapi_base",
		"demo_state",
	)
	fmt.Println("OAuth 授权链接:")
	fmt.Println(authURL)

	// 可选：如果你已经从回调拿到了 code，这里演示如何在服务端换取 token/openid
	if code := strings.TrimSpace(os.Getenv("WECHAT_TEST_CODE")); code != "" {
		app := cli
		tokenResp, err := app.OAuth.TokenFromCode(code)
		if err != nil {
			log.Fatalf("TokenFromCode error: %v", err)
		}
		// tokenResp 是 *object.HashMap，包含 access_token、openid 等字段
		fmt.Printf("TokenFromCode 响应: %#v\n", tokenResp)
	}

	// 4) 可选：发送一条客服文本消息（需要 OPENID 和 文本内容）
	openID := os.Getenv("WECHAT_TEST_OPENID")
	text := os.Getenv("WECHAT_TEST_TEXT")
	if openID != "" && text != "" {
		fmt.Println("尝试发送客服消息…")
		// 使用 PowerWeChat 直接发送客服文本消息，参考：https://powerwechat.artisan-cloud.com/zh/official-account/messages.html
		app := cli
		msg := messages.NewText(text)
		if _, err := app.CustomerService.Message(context.Background(), msg).SetTo(openID).Send(context.Background()); err != nil {
			log.Fatalf("send text error: %v", err)
		}
		fmt.Println("消息已发送")
	} else {
		fmt.Println("未设置 WECHAT_TEST_OPENID/WECHAT_TEST_TEXT，跳过发送客服消息演示")
	}
}
