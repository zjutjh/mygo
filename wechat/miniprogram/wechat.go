package miniprogram

import (
	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/response"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

type WeChat struct {
	conf        Config
	miniProgram *miniProgram.MiniProgram
}

/**
 * 使用方法请参考官方文档
 * 入门://powerwechat.artisan-cloud.com/zh/mini-program/
 * 小程序登录：https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html
 * 获取用户信息：https://powerwechat.artisan-cloud.com/zh/mini-program/user-info.html
 */

// New 创建微信服务实例
func New(conf Config) *WeChat {

	var kernelCache cache.CacheInterface
	gr := cache.NewGRedis(&redis.UniversalOptions{})
	gr.Pool = nedis.Pick(conf.Redis)
	kernelCache = gr

	mp, _ := miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:        conf.AppId,
		Secret:       conf.AppSecret,
		ResponseType: response.TYPE_MAP,
		Token:        conf.Token,
		AESKey:       conf.AesKey,

		HttpDebug: conf.HttpDebug,
		Log: miniProgram.Log{
			Level:  conf.Log.Level,
			File:   conf.Log.File,
			Error:  conf.Log.Error,
			Stdout: conf.Log.Stdout,
		},

		Cache: kernelCache,
	})

	return &WeChat{
		conf:        conf,
		miniProgram: mp,
	}
}
