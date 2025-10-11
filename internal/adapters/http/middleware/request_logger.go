package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"library-service/pkg/logutil"
)

// responseWriter is a wrapper to capture response status and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// RequestLogger middleware logs all HTTP requests and responses
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context
			requestID := logutil.GetRequestID(r.Context())

			// Create context logger with request metadata
			contextLogger := logger.With(
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)

			// Add logger to context for downstream use
			ctx := logutil.WithLogger(r.Context(), contextLogger)

			// Log request body for debugging (limit size to prevent memory issues)
			if r.Method != http.MethodGet && r.ContentLength > 0 && r.ContentLength < 10240 { // 10KB limit
				body, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
				contextLogger = contextLogger.With(zap.ByteString("request_body", body))
			}

			// Log incoming request
			contextLogger.Info("incoming request",
				zap.String("user_agent", r.UserAgent()),
				zap.Int64("content_length", r.ContentLength),
			)

			// Wrap response writer to capture status and size
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Process request with enriched context
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log completed request
			duration := time.Since(start)
			contextLogger.Info("request completed",
				zap.Int("status", wrapped.status),
				zap.Int("response_size", wrapped.size),
				zap.Duration("duration", duration),
				zap.String("duration_ms", duration.Truncate(time.Millisecond).String()),
			)

			// Slow request warning
			if duration > 1*time.Second {
				contextLogger.Warn("slow request detected",
					zap.Duration("duration", duration),
				)
			}
		})
	}
}

// ContextLogger returns a logger with request context
func ContextLogger(r *http.Request) *zap.Logger {
	return logutil.FromContext(r.Context())
}
