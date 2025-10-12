package logutil

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Context keys for logging
type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	traceIDKey   contextKey = "trace_id"
)

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	// Return default logger if none in context
	return zap.L()
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, requestIDKey, requestID)

	// Also add to logger if present
	if logger := FromContext(ctx); logger != nil {
		logger = logger.With(zap.String("request_id", requestID))
		ctx = WithLogger(ctx, logger)
	}

	return ctx
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// GenerateRequestID generates a new request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// WithTraceID adds trace ID to context for distributed tracing
func WithTraceID(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)

	// Also add to logger if present
	if logger := FromContext(ctx); logger != nil {
		logger = logger.With(zap.String("trace_id", traceID))
		ctx = WithLogger(ctx, logger)
	}

	return ctx
}

// GetTraceID retrieves trace ID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)

	// Also add to logger if present
	if logger := FromContext(ctx); logger != nil {
		logger = logger.With(zap.String("user_id", userID))
		ctx = WithLogger(ctx, logger)
	}

	return ctx
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// EnrichContext adds common fields to context and logger
func EnrichContext(ctx context.Context, fields map[string]interface{}) context.Context {
	logger := FromContext(ctx)

	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))

		// Special handling for certain keys
		switch key {
		case "request_id":
			if id, ok := value.(string); ok {
				ctx = context.WithValue(ctx, requestIDKey, id)
			}
		case "user_id", "member_id":
			if id, ok := value.(string); ok {
				ctx = context.WithValue(ctx, userIDKey, id)
			}
		case "trace_id":
			if id, ok := value.(string); ok {
				ctx = context.WithValue(ctx, traceIDKey, id)
			}
		}
	}

	logger = logger.With(zapFields...)
	return WithLogger(ctx, logger)
}

// PropagateContext creates a new context with propagated values
func PropagateContext(parent, child context.Context) context.Context {
	// Propagate request ID
	if requestID := GetRequestID(parent); requestID != "" {
		child = WithRequestID(child, requestID)
	}

	// Propagate trace ID
	if traceID := GetTraceID(parent); traceID != "" {
		child = WithTraceID(child, traceID)
	}

	// Propagate user ID
	if userID := GetUserID(parent); userID != "" {
		child = WithUserID(child, userID)
	}

	// Propagate logger
	if logger := FromContext(parent); logger != nil {
		child = WithLogger(child, logger)
	}

	return child
}

// ContextFields extracts all logging fields from context
func ContextFields(ctx context.Context) []zap.Field {
	fields := []zap.Field{}

	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// Check for member_id (might be set by auth middleware)
	if memberID, ok := ctx.Value("member_id").(string); ok && memberID != "" {
		fields = append(fields, zap.String("member_id", memberID))
	}

	// Check for role (might be set by auth middleware)
	if role, ok := ctx.Value("role").(string); ok && role != "" {
		fields = append(fields, zap.String("role", role))
	}

	return fields
}

// StartOperation starts a new operation with logging
func StartOperation(ctx context.Context, operation string) (context.Context, func()) {
	logger := FromContext(ctx).Named(operation)

	// Generate request ID if not present
	if GetRequestID(ctx) == "" {
		ctx = WithRequestID(ctx, GenerateRequestID())
	}

	// Add operation to logger
	fields := append([]zap.Field{zap.String("operation", operation)}, ContextFields(ctx)...)
	logger = logger.With(fields...)

	ctx = WithLogger(ctx, logger)

	logger.Debug("operation started")

	// Return cleanup function
	return ctx, func() {
		logger.Debug("operation completed")
	}
}

// StartSpan starts a new span for tracing
func StartSpan(ctx context.Context, spanName string) (context.Context, func()) {
	logger := FromContext(ctx)

	// Generate trace ID if not present
	if GetTraceID(ctx) == "" {
		ctx = WithTraceID(ctx, GenerateRequestID())
	}

	spanID := GenerateRequestID()[:8] // Short span ID

	logger = logger.With(
		zap.String("span_name", spanName),
		zap.String("span_id", spanID),
	)

	ctx = WithLogger(ctx, logger)

	logger.Debug("span started", zap.String("span", spanName))

	// Return cleanup function
	return ctx, func() {
		logger.Debug("span completed", zap.String("span", spanName))
	}
}
