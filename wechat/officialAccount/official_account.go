package officialAccount

import (
	"fmt"

	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

/*
关于Oauth和messages方法，请参考官方文档：
入门https://powerwechat.artisan-cloud.com/zh/official-account/
消息https://powerwechat.artisan-cloud.com/zh/official-account/messages.html
网页授权https://powerwechat.artisan-cloud.com/zh/official-account/oauth.html

*/

func New(conf Config) (*officialAccount.OfficialAccount, error) {
	var kernelCache cache.CacheInterface
	switch conf.Driver {
	case DriverRedis:
		gr := cache.NewGRedis(&redis.UniversalOptions{})
		gr.Pool = nedis.Pick(conf.Redis)
		kernelCache = gr
	case DriverMemory:
		kernelCache = cache.NewMemCache(conf.MemCache.Namespace, conf.MemCache.DefaultExpire, conf.MemCache.Prefix)
	default:
		kernelCache = cache.NewMemCache(conf.MemCache.Namespace, conf.MemCache.DefaultExpire, conf.MemCache.Prefix)
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
		Cache: kernelCache,
	}

	app, err := officialAccount.NewOfficialAccount(pwConf)
	if err != nil {
		return nil, fmt.Errorf("初始化 PowerWeChat OfficialAccount 失败: %w", err)
	}

	return app, nil
}
