package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"

	"library-service/pkg/logutil"
)

// Recovery middleware recovers from panics and returns a proper error response
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get request ID for correlation
					requestID := logutil.GetRequestID(r.Context())

					// Get stack trace
					stack := debug.Stack()

					// Log the panic with full context
					logger.Error("panic recovered",
						zap.String("request_id", requestID),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Any("panic", err),
						zap.ByteString("stack", stack),
					)

					// Return error response to client
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Request-ID", requestID)
					w.WriteHeader(http.StatusInternalServerError)

					response := map[string]interface{}{
						"error":      "Internal server error",
						"message":    "An unexpected error occurred",
						"request_id": requestID,
						"code":       http.StatusInternalServerError,
					}

					// In development mode, include panic details
					if isDevelopment() {
						response["panic"] = fmt.Sprintf("%v", err)
						response["stack"] = string(stack)
					}

					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// isDevelopment checks if we're in development mode
func isDevelopment() bool {
	// This could be read from environment or config
	// For now, simple check
	return false
}
