package officalAccount

import (
	"fmt"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
)

// OfficalAccount 封装 PowerWeChat 公众号客户端
type OfficalAccount struct {
	conf Config
	app  *officialAccount.OfficialAccount
}

// New 基于配置初始化 PowerWeChat 公众号客户端
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

// Client 返回底层 PowerWeChat OfficialAccount 客户端
func (o *OfficalAccount) Client() *officialAccount.OfficialAccount {
	return o.app
}
