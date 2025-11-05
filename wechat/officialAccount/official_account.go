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

type OfficalAccount struct {
	conf Config
	app  *officialAccount.OfficialAccount
}

func New(conf Config) (*OfficalAccount, error) {
	if !conf.Enable {
		return nil, fmt.Errorf("official account 未启用")
	}

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

	return &OfficalAccount{
		conf: conf,
		app:  app,
	}, nil
}

func (o *OfficalAccount) Client() *officialAccount.OfficialAccount {
	return o.app
}
