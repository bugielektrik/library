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

// WithLogger returns a new context containing the provided logger.
// Use this to pass a logger down call chains for structured logging.
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns a logger stored in the context or the package default logger.
// Always returns a non-nil *zap.Logger.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return GetLogger()
	}
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return GetLogger()
}

// GetLogger returns the singleton default logger, initializing it on first use.
// The logger returned should be Sync()'d by the application on shutdown.
func GetLogger() *zap.Logger {
	once.Do(func() {
		if err := initDefaultLogger(); err != nil {
			// fallback to an example logger if initialization fails
			fallback := zap.NewExample()
			fallback.Warn("failed to initialize logger, using fallback", zap.Error(err))
			defaultLogger = fallback
		}
	})
	// Ensure we never return nil
	if defaultLogger == nil {
		defaultLogger = zap.NewNop()
	}
	return defaultLogger
}

// NewLogger builds a zap.Logger according to environment configuration and returns it.
// Caller is responsible for calling Sync() at shutdown.
// Returns an error if building the logger fails.
func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	// use development config when not in prod
	if os.Getenv("APP_MODE") != "prod" {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	// structured, lowercase keys and ISO8601 timestamps
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// outputs: stdout and service.log
	cfg.OutputPaths = []string{"stdout", "service.log"}

	// wrap core with apmzap for APM integration; ensure a reasonable flush timeout
	apmCore := &apmzap.Core{FatalFlushTimeout: 10 * time.Second}
	logger, err := cfg.Build(zap.WrapCore(apmCore.WrapCore))
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// New is a convenience wrapper that returns a logger and guarantees a usable logger
// even if building the configured logger fails. It logs a warning in that case.
// Prefer NewLogger when you want explicit error handling.
func New() *zap.Logger {
	l, err := NewLogger()
	if err != nil {
		fallback := zap.NewExample()
		// structured, lowercase message with params
		fallback.Warn("unable to build configured logger, using example fallback", zap.Error(err))
		return fallback
	}
	return l
}

// initDefaultLogger initializes the package-level default logger.
func initDefaultLogger() error {
	l, err := NewLogger()
	if err != nil {
		return err
	}
	defaultLogger = l
	return nil
}

// SyncLogger flushes any buffered log entries for the provided logger.
// Many environments may return an error when syncing (e.g., on Windows); callers
// may choose to ignore it. This helper centralizes that call.
func SyncLogger(l *zap.Logger) error {
	if l == nil {
		return nil
	}
	return l.Sync()
}
