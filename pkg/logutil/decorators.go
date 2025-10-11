package logutil

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"go.uber.org/zap"
)

// Decorator wraps a function with automatic logging
type Decorator struct {
	logger    *zap.Logger
	threshold time.Duration // Log warning if execution takes longer
}

// NewDecorator creates a new logging decorator
func NewDecorator(logger *zap.Logger) *Decorator {
	return &Decorator{
		logger:    logger,
		threshold: time.Second, // Default 1 second threshold
	}
}

// WithThreshold sets the performance warning threshold
func (d *Decorator) WithThreshold(threshold time.Duration) *Decorator {
	d.threshold = threshold
	return d
}

// LogExecution wraps a function with execution logging
func (d *Decorator) LogExecution(fn interface{}) interface{} {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		panic("LogExecution: argument must be a function")
	}

	// Get function name for logging
	fnName := runtime.FuncForPC(fnValue.Pointer()).Name()

	// Create wrapper function
	wrapper := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		start := time.Now()

		// Extract context if present
		ctx := extractContext(args)
		logger := d.enrichLogger(ctx)

		// Log entry
		logger.Debug("function entry",
			zap.String("function", fnName),
			zap.Int("args_count", len(args)),
		)

		// Call original function
		results := fnValue.Call(args)

		// Check for errors
		duration := time.Since(start)
		if err := extractError(results); err != nil {
			logger.Error("function failed",
				zap.String("function", fnName),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		} else if duration > d.threshold {
			logger.Warn("slow function execution",
				zap.String("function", fnName),
				zap.Duration("duration", duration),
				zap.Duration("threshold", d.threshold),
			)
		} else {
			logger.Debug("function exit",
				zap.String("function", fnName),
				zap.Duration("duration", duration),
			)
		}

		return results
	})

	return wrapper.Interface()
}

// LogMethod wraps a method with automatic logging and error handling
func LogMethod(ctx context.Context, operation string, fn func() error) error {
	logger := FromContext(ctx)
	start := time.Now()

	logger.Debug("method entry",
		zap.String("operation", operation),
	)

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error("method failed",
			zap.String("operation", operation),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
	} else {
		logger.Debug("method completed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// LogMethodWithResult wraps a method that returns a result and error
func LogMethodWithResult[T any](ctx context.Context, operation string, fn func() (T, error)) (T, error) {
	logger := FromContext(ctx)
	start := time.Now()

	logger.Debug("method entry",
		zap.String("operation", operation),
	)

	result, err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error("method failed",
			zap.String("operation", operation),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
	} else {
		logger.Debug("method completed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
		)
	}

	return result, err
}

// UseCase creates a specialized logger for use case operations
func UseCase(ctx context.Context, name string) (*zap.Logger, context.Context) {
	logger := FromContext(ctx).Named(name)

	// Add use case metadata
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Add to context
	ctx = WithLogger(ctx, logger)

	return logger, ctx
}

// UseCaseLogger creates a logger for a specific use case
func UseCaseLogger(ctx context.Context, useCase, operation string) *zap.Logger {
	logger := FromContext(ctx).Named(fmt.Sprintf("%s_usecase", useCase))

	logger = logger.With(
		zap.String("use_case", useCase),
		zap.String("operation", operation),
	)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Add member ID if available
	if memberID, ok := ctx.Value("member_id").(string); ok && memberID != "" {
		logger = logger.With(zap.String("member_id", memberID))
	}

	return logger
}

// HandlerLogger creates a logger for HTTP handlers
func HandlerLogger(ctx context.Context, handler, operation string) *zap.Logger {
	logger := FromContext(ctx).Named(fmt.Sprintf("%s_handler", handler))

	logger = logger.With(
		zap.String("handler", handler),
		zap.String("operation", operation),
	)

	// Add request metadata
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Add member ID if available
	if memberID, ok := ctx.Value("member_id").(string); ok && memberID != "" {
		logger = logger.With(zap.String("member_id", memberID))
	}

	// Add role if available
	if role, ok := ctx.Value("role").(string); ok && role != "" {
		logger = logger.With(zap.String("role", role))
	}

	return logger
}

// RepositoryLogger creates a logger for repository operations
func RepositoryLogger(ctx context.Context, repository, operation string) *zap.Logger {
	logger := FromContext(ctx).Named(fmt.Sprintf("%s_repository", repository))

	logger = logger.With(
		zap.String("repository", repository),
		zap.String("operation", operation),
		zap.String("layer", "repository"),
	)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	return logger
}

// ServiceLogger creates a logger for domain service operations
func ServiceLogger(ctx context.Context, service, operation string) *zap.Logger {
	logger := FromContext(ctx).Named(fmt.Sprintf("%s_service", service))

	logger = logger.With(
		zap.String("service", service),
		zap.String("operation", operation),
		zap.String("layer", "domain"),
	)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	return logger
}

// extractContext finds and returns context from function arguments
func extractContext(args []reflect.Value) context.Context {
	for _, arg := range args {
		if ctx, ok := arg.Interface().(context.Context); ok {
			return ctx
		}
	}
	return context.Background()
}

// extractError finds and returns error from function results
func extractError(results []reflect.Value) error {
	for _, result := range results {
		if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !result.IsNil() {
				return result.Interface().(error)
			}
		}
	}
	return nil
}

// enrichLogger adds context information to logger
func (d *Decorator) enrichLogger(ctx context.Context) *zap.Logger {
	logger := d.logger

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	// Add member ID if available
	if memberID, ok := ctx.Value("member_id").(string); ok && memberID != "" {
		logger = logger.With(zap.String("member_id", memberID))
	}

	return logger
}

// LogDatabaseQuery logs database query execution
func LogDatabaseQuery(ctx context.Context, query string, duration time.Duration, err error) {
	logger := FromContext(ctx).Named("database")

	fields := []zap.Field{
		zap.String("query", truncateQuery(query)),
		zap.Duration("duration", duration),
	}

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if err != nil {
		logger.Error("query failed", append(fields, zap.Error(err))...)
	} else if duration > 100*time.Millisecond {
		logger.Warn("slow query", fields...)
	} else {
		logger.Debug("query executed", fields...)
	}
}

// LogCacheOperation logs cache operations
func LogCacheOperation(ctx context.Context, operation string, key string, hit bool, duration time.Duration) {
	logger := FromContext(ctx).Named("cache")

	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.Duration("duration", duration),
	}

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	logger.Debug("cache operation", fields...)
}

// truncateQuery truncates long queries for logging
func truncateQuery(query string) string {
	const maxLength = 200
	if len(query) <= maxLength {
		return query
	}
	return query[:maxLength] + "..."
}

// LogError logs an error with context
func LogError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	logger := FromContext(ctx)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// Add error fields
	fields = append(fields, zap.Error(err))

	// Check if it's a domain error with details
	if domainErr, ok := err.(interface {
		GetDetail(string) (interface{}, bool)
	}); ok {
		if requestID, ok := domainErr.GetDetail("request_id"); ok {
			fields = append(fields, zap.Any("error_request_id", requestID))
		}
		if cause, ok := domainErr.GetDetail("cause"); ok {
			fields = append(fields, zap.Any("error_cause", cause))
		}
	}

	logger.Error(msg, fields...)
}

// LogInfo logs an informational message with context
func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	logger := FromContext(ctx)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	logger.Info(msg, fields...)
}

// LogDebug logs a debug message with context
func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	logger := FromContext(ctx)

	// Add request ID if available
	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	logger.Debug(msg, fields...)
}
