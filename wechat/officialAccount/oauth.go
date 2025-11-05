package officialAccount

import (
	"fmt"
	"net/url"

	"github.com/ArtisanCloud/PowerLibs/v3/object"
)

// BuildOAuthURL 构造微信网页授权跳转链接
// scope 可选: snsapi_base / snsapi_userinfo
// state 可为空
func (o *OfficalAccount) BuildOAuthURL(redirectURI, scope, state string) (string, error) {
	if redirectURI == "" {
		return "", fmt.Errorf("redirectURI 不能为空")
	}
	if scope == "" {
		scope = "snsapi_base"
	}
	esc := url.QueryEscape(redirectURI)
	// 标准 OAuth2 授权 URL
	u := fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		o.conf.AppID,
		esc,
		scope,
		url.QueryEscape(state),
	)
	return u, nil
}

// OAuthTokenFromCode 通过 code 换取网页授权 access_token/openid 等
func (o *OfficalAccount) OAuthTokenFromCode(code string) (*object.HashMap, error) {
	if code == "" {
		return nil, fmt.Errorf("code 不能为空")
	}
	return o.app.OAuth.TokenFromCode(code)
}

// OAuthUserFromCode 通过 code 直接获取用户信息（需 scope=snsapi_userinfo）
func (o *OfficalAccount) OAuthUserFromCode(code string) (any, error) {
	if code == "" {
		return nil, fmt.Errorf("code 不能为空")
	}
	return o.app.OAuth.UserFromCode(code)
}

// OAuthUserFromToken 通过 accessToken 和 openID 获取用户信息
func (o *OfficalAccount) OAuthUserFromToken(accessToken, openID string) (any, error) {
	if accessToken == "" || openID == "" {
		return nil, fmt.Errorf("accessToken 和 openID 不能为空")
	}
	return o.app.OAuth.UserFromToken(accessToken, openID)
}
