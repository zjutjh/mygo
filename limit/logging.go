package limit

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type loggingLimiter struct {
	next   Limiter
	logger *logrus.Logger
}

func (l *loggingLimiter) Allow(ctx context.Context, key string, limit int, burst int) (bool, time.Duration, error) {
	allowed, retry, err := l.next.Allow(ctx, key, limit, burst)
	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"module": "limit",
			"key":    key,
			"error":  err,
		}).Errorf("limit: 检查失败 [key: %s]", key) // 错误级别日志
		return allowed, retry, err
	}

	if !allowed {
		l.logger.WithFields(logrus.Fields{
			"module":      "limit",
			"key":         key,
			"limit":       limit,
			"burst":       burst,
			"retry_after": retry.String(),
		}).Warnf("limit: 请求超出速率限制 [key: %s]", key) // 警告级别日志
	}

	return allowed, retry, nil
}
