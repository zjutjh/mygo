package nmdb

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type mdbLogger struct {
	logLevel                  LogLevel
	slowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	Colorful                  bool
	l                         *logrus.Logger
}

func newDBLogger(l *logrus.Logger, conf Config) *mdbLogger {
	return &mdbLogger{
		logLevel:                  conf.LogLevel,
		slowThreshold:             conf.SlowThreshold,
		IgnoreRecordNotFoundError: conf.IgnoreRecordNotFoundError,
		Colorful:                  conf.Colorful,
		l:                         l,
	}
}

// LogMode log mode
func (l *mdbLogger) LogMode(level LogLevel) *mdbLogger {
	nl := *l
	nl.logLevel = level
	return &nl
}

// Info print info
func (l *mdbLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= Info {
		l.l.WithContext(ctx).WithField("line", getCallerInfo()).Infof(msg, data...)
	}
}

// Warn print warn messages
func (l *mdbLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= Warn {
		l.l.WithContext(ctx).WithField("line", getCallerInfo()).Warnf(msg, data...)
	}
}

// Error print error messages
func (l *mdbLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= Error {
		l.l.WithContext(ctx).WithField("line", getCallerInfo()).Errorf(msg, data...)
	}
}

// Trace print sql message
func (l *mdbLogger) Trace(ctx context.Context, command string, duration time.Duration,
	collection string, operation string, resultCount int64, err error) {

	if l.logLevel <= Silent {
		return
	}

	switch {
	case err != nil && l.logLevel >= Error && (!errors.Is(err, mongo.ErrNoDocuments) || !l.IgnoreRecordNotFoundError):

		l.l.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"line":       getCallerInfo(),
			"cost":       duration.String(),
			"collection": collection,
			"operation":  operation,
			"command":    command,
		}).Error("MongoDB 命令执行错误")
	case duration > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= Warn:
		l.l.WithContext(ctx).WithFields(logrus.Fields{
			"line":       getCallerInfo(),
			"cost":       duration.String(),
			"collection": collection,
			"operation":  operation,
			"threshold":  l.slowThreshold.String(),
		}).Warn("MongoDB 操作时长超过期望")
	case l.logLevel == Info:
		l.l.WithContext(ctx).WithFields(logrus.Fields{
			"line":        getCallerInfo(),
			"cost":        duration.String(),
			"collection":  collection,
			"operation":   operation,
			"resultCount": resultCount,
		}).Info("MongoDB 操作记录")
	}

}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:0"
	}
	return filepath.Base(file) + ":" + fmt.Sprintf("%d", line)
}
