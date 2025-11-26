package log

import (
	"context"
	"os"
	"sync"
	"time"

	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey string

const loggerKey ctxKey = "logger"

var (
	defaultLogger *zap.Logger
	once          sync.Once
)

func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return GetLogger()
	}
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return GetLogger()
}

func GetLogger() *zap.Logger {
	once.Do(func() {
		if err := initDefaultLogger(); err != nil {
			fallback := zap.NewExample()
			fallback.Warn("failed to initialize logger, using fallback", zap.Error(err))
			defaultLogger = fallback
		}
	})
	if defaultLogger == nil {
		defaultLogger = zap.NewNop()
	}
	return defaultLogger
}

func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	if os.Getenv("APP_MODE") != "prod" {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg.OutputPaths = []string{"stdout", "service.log"}

	apmCore := &apmzap.Core{FatalFlushTimeout: 10 * time.Second}
	logger, err := cfg.Build(zap.WrapCore(apmCore.WrapCore))
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func New() *zap.Logger {
	l, err := NewLogger()
	if err != nil {
		fallback := zap.NewExample()
		fallback.Warn("unable to build configured logger, using example fallback", zap.Error(err))
		return fallback
	}
	return l
}

func initDefaultLogger() error {
	l, err := NewLogger()
	if err != nil {
		return err
	}
	defaultLogger = l
	return nil
}

func SyncLogger(l *zap.Logger) error {
	if l == nil {
		return nil
	}
	return l.Sync()
}
