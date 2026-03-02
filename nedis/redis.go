package nedis

import (
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"

	"github.com/zjutjh/mygo/nlog"
)

// New 以指定配置创建实例
func New(conf Config) redis.UniversalClient {
	// 选中logger
	l := nlog.Pick(conf.Log)

	// 创建hook
	h := &hook{
		logger:         l,
		InfoRecordTime: conf.InfoRecordTime,
		WarnRecordTime: conf.WarnRecordTime,
	}

	// 初始化options
	options := &redis.UniversalOptions{
		Addrs:                 conf.Addrs,
		ClientName:            conf.ClientName,
		DB:                    conf.DB,
		Protocol:              conf.Protocol,
		Username:              conf.Username,
		Password:              conf.Password,
		SentinelUsername:      conf.SentinelUsername,
		SentinelPassword:      conf.SentinelPassword,
		MaxRetries:            conf.MaxRetries,
		MinRetryBackoff:       conf.MinRetryBackoff,
		MaxRetryBackoff:       conf.MaxRetryBackoff,
		DialTimeout:           conf.DialTimeout,
		ReadTimeout:           conf.ReadTimeout,
		WriteTimeout:          conf.WriteTimeout,
		ContextTimeoutEnabled: conf.ContextTimeoutEnabled,
		ReadBufferSize:        conf.ReadBufferSize,
		WriteBufferSize:       conf.WriteBufferSize,
		PoolFIFO:              conf.PoolFIFO,
		PoolSize:              conf.PoolSize,
		PoolTimeout:           conf.PoolTimeout,
		MinIdleConns:          conf.MinIdleConns,
		MaxIdleConns:          conf.MaxIdleConns,
		MaxActiveConns:        conf.MaxActiveConns,
		ConnMaxIdleTime:       conf.ConnMaxIdleTime,
		ConnMaxLifetime:       conf.ConnMaxLifetime,
		MaxRedirects:          conf.MaxRedirects,
		ReadOnly:              conf.ReadOnly,
		RouteByLatency:        conf.RouteByLatency,
		RouteRandomly:         conf.RouteRandomly,
		MasterName:            conf.MasterName,
		DisableIdentity:       conf.DisableIdentity,
		IdentitySuffix:        conf.IdentitySuffix,
		FailingTimeoutSeconds: conf.FailingTimeoutSeconds,
		UnstableResp3:         conf.UnstableResp3,
		IsClusterMode:         conf.IsClusterMode,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode:                           conf.MaintNotificationsConfig.Mode,
			EndpointType:                   conf.MaintNotificationsConfig.EndpointType,
			RelaxedTimeout:                 conf.MaintNotificationsConfig.RelaxedTimeout,
			HandoffTimeout:                 conf.MaintNotificationsConfig.HandoffTimeout,
			MaxWorkers:                     conf.MaintNotificationsConfig.MaxWorkers,
			HandoffQueueSize:               conf.MaintNotificationsConfig.HandoffQueueSize,
			PostHandoffRelaxedDuration:     conf.MaintNotificationsConfig.PostHandoffRelaxedDuration,
			CircuitBreakerFailureThreshold: conf.MaintNotificationsConfig.CircuitBreakerFailureThreshold,
			CircuitBreakerResetTimeout:     conf.MaintNotificationsConfig.CircuitBreakerResetTimeout,
			CircuitBreakerMaxRequests:      conf.MaintNotificationsConfig.CircuitBreakerMaxRequests,
			MaxHandoffRetries:              conf.MaintNotificationsConfig.MaxHandoffRetries,
		},
	}

	// 创建client并挂载hook
	var client redis.UniversalClient
	switch conf.Mode {
	case ModeCluster:
		client = redis.NewClusterClient(options.Cluster())
	case ModeFailover:
		client = redis.NewFailoverClient(options.Failover())
	case ModeSingle:
		client = redis.NewClient(options.Simple())
	default:
		client = redis.NewClient(options.Simple())
	}
	client.AddHook(h)

	return client
}
