package log

import (
	"context"
	"os"

	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger *zap.Logger
)

func init() {
	defaultLogger = New()
}

type logger struct{}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logger{}, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(logger{}).(*zap.Logger); ok {
		return l
	}
	lg := defaultLogger

	return lg
}

func New() *zap.Logger {
	cfg := zap.NewProductionConfig()

	if os.Getenv("DEBUG") != "" {
		cfg = zap.NewDevelopmentConfig()

		if os.Getenv("DEBUG") == "true" {
			cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		}
	}

	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"stdout", "service.log"}

	log, err := cfg.Build(zap.WrapCore((&apmzap.Core{FatalFlushTimeout: 10000}).WrapCore))
	if err != nil {
		log = zap.NewExample()
		log.Warn("Unable to set up the logger. Replaced with example one which shouldn't fail", zap.Error(err))
	}
	defer log.Sync()

	return log
}
