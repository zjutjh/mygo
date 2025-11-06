package officialAccount

import (
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
)

/*
关于Oauth和messages方法，请参考官方文档：
入门https://powerwechat.artisan-cloud.com/zh/official-account/
消息https://powerwechat.artisan-cloud.com/zh/official-account/messages.html
网页授权https://powerwechat.artisan-cloud.com/zh/official-account/oauth.html

*/

// New 直接返回 PowerWeChat 的 OfficialAccount 客户端实例
func New(conf Config) (*officialAccount.OfficialAccount, error) {

	pwConf := &officialAccount.UserConfig{
		AppID:     conf.AppID,
		Secret:    conf.Secret,
		Token:     conf.Token,
		AESKey:    conf.AESKey,
		HttpDebug: conf.HttpDebug,
		Debug:     conf.Debug,
		Log: officialAccount.Log{
			Level:  conf.Log.Level,
			File:   conf.Log.File,
			Error:  conf.Log.Error,
			Stdout: conf.Log.Stdout,
		},
	}

	app, err := officialAccount.NewOfficialAccount(pwConf)
	if err != nil {
		return nil, fmt.Errorf("初始化 PowerWeChat OfficialAccount 失败: %w", err)
	}

	return app, nil
}
