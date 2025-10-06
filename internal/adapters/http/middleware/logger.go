package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"library-service/internal/infrastructure/log"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// RequestLogger logs HTTP requests with timing and response information
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Add logger to context
			ctx := log.WithLogger(r.Context(), logger)
			r = r.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", wrapped.status),
				zap.Int("size", wrapped.size),
				zap.Duration("duration", duration),
				zap.String("ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			// Use different log levels based on status code
			switch {
			case wrapped.status >= 500:
				logger.Error("request completed with error", fields...)
			case wrapped.status >= 400:
				logger.Warn("request completed with client error", fields...)
			default:
				logger.Info("request completed", fields...)
			}
		})
	}
}
