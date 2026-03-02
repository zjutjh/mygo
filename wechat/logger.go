package wechat

import (
	"context"
	"fmt"

	"github.com/ArtisanCloud/PowerLibs/v3/logger/contract"
	"github.com/sirupsen/logrus"
)

type wechatLogger struct {
	logger  *logrus.Logger
	context context.Context
}

func NewLogger(l *logrus.Logger) contract.LoggerInterface {
	return &wechatLogger{logger: l}
}

func (l *wechatLogger) Debug(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Debug(msg)
}

func (l *wechatLogger) Info(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Info(msg)
}

func (l *wechatLogger) Warn(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Warn(msg)
}

func (l *wechatLogger) Error(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Error(msg)
}

func (l *wechatLogger) Panic(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Panic(msg)
}

func (l *wechatLogger) Fatal(msg string, v ...any) {
	l.logger.WithContext(l.context).WithFields(l.sweetenFields(v...)).Fatal(msg)
}

func (l *wechatLogger) DebugF(format string, args ...any) {
	l.logger.WithContext(l.context).Debugf(format, args...)
}

func (l *wechatLogger) InfoF(format string, args ...any) {
	l.logger.WithContext(l.context).Infof(format, args...)
}

func (l *wechatLogger) WarnF(format string, args ...any) {
	l.logger.WithContext(l.context).Warnf(format, args...)
}

func (l *wechatLogger) ErrorF(format string, args ...any) {
	l.logger.WithContext(l.context).Errorf(format, args...)
}

func (l *wechatLogger) PanicF(format string, args ...any) {
	l.logger.WithContext(l.context).Panicf(format, args...)
}

func (l *wechatLogger) FatalF(format string, args ...any) {
	l.logger.WithContext(l.context).Fatalf(format, args...)
}

func (l *wechatLogger) WithContext(ctx context.Context) contract.LoggerInterface {
	l.context = ctx
	return l
}

func (l *wechatLogger) sweetenFields(keysAndValues ...any) logrus.Fields {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make(logrus.Fields, len(keysAndValues)/2)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			l.logger.WithContext(l.context).Warnf("Found dangling key at index %d, ignoring: %#v\n", i, keysAndValues[i])
			break
		}

		key := keysAndValues[i]
		val := keysAndValues[i+1]

		keyStr, ok := key.(string)
		if !ok {
			placeholderKey := fmt.Sprintf("non_string_key_at_index_%d", i)
			l.logger.WithContext(l.context).Warnf("Non-string key found at index %d (%#v). Using key: %s\n", i, key, placeholderKey)
			fields[placeholderKey] = val
			continue
		}

		fields[keyStr] = val
	}

	return fields
}
