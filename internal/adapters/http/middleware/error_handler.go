package middleware

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	log "library-service/internal/infrastructure/logger"
	"library-service/pkg/errors"
)

// ErrorHandler is a middleware that recovers from panics and handles errors
func ErrorHandler(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)

					respondError(w, r, errors.ErrInternal.Wrap(err.(error)))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// respondError writes an error response
func respondError(w http.ResponseWriter, r *http.Request, err error) {
	logger := log.FromContext(r.Context())

	// Determine HTTP status code
	status := errors.GetHTTPStatus(err)

	// Log the error
	if status >= 500 {
		logger.Error("internal error",
			zap.Error(err),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
		)
	} else {
		logger.Warn("client error",
			zap.Error(err),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Int("status", status),
		)
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.FromError(err)
	json.NewEncoder(w).Encode(response)
}
