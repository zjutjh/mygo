package miniProgram

import (
	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/redis/go-redis/v9"

	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/nesty"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/wechat"
)

// New 以指定配置创建实例
// https://powerwechat.artisan-cloud.com/zh/mini-program/
func New(conf Config) (*miniProgram.MiniProgram, error) {
	l := nlog.Pick(conf.Log)
	client := nesty.Pick(conf.Resty)
	rdb := nedis.Pick(conf.Redis)

	var kernelCache cache.CacheInterface
	switch conf.Cache.Driver {
	case CacheDriverRedis:
		gr := cache.NewGRedis(&redis.UniversalOptions{})
		gr.Pool = rdb
		kernelCache = gr
	case CacheDriverMemory:
		kernelCache = cache.NewMemCache(conf.Cache.MemCache.Namespace, conf.Cache.MemCache.DefaultLifeTime, conf.Cache.MemCache.Prefix)
	default:
		kernelCache = nil
	}

	uc := &miniProgram.UserConfig{
		AppID:             conf.AppID,
		Secret:            conf.Secret,
		AppKey:            conf.AppKey,
		OfferID:           conf.OfferID,
		Token:             conf.OfferID,
		AESKey:            conf.AESKey,
		ComponentAppID:    conf.ComponentAppID,
		ComponentAppToken: conf.ComponentAppToken,
		StableTokenMode:   conf.StableTokenMode,
		RefreshToken:      conf.RefreshToken,
		ResponseType:      conf.ResponseType,
		Http: miniProgram.Http{
			Timeout:   conf.Http.Timeout,
			BaseURI:   conf.Http.BaseURI,
			ProxyURI:  conf.Http.ProxyURI,
			Transport: client.GetClient().Transport,
		},
		Log: miniProgram.Log{
			Driver: wechat.NewLogger(l),
		},
		Cache:     kernelCache,
		HttpDebug: conf.HttpDebug,
		Debug:     conf.Debug,
	}

	mp, err := miniProgram.NewMiniProgram(uc)
	if err != nil {
		return nil, err
	}

	return mp, nil
}
