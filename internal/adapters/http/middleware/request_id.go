package middleware

import (
	"net/http"

	"library-service/pkg/logutil"
)

// RequestIDHeader is the header name for request ID
const RequestIDHeader = "X-Request-ID"

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID (from client or load balancer)
		requestID := r.Header.Get(RequestIDHeader)

		// Generate new ID if not present
		if requestID == "" {
			requestID = logutil.GenerateRequestID()
		}

		// Add to context using logutil (which also adds to logger)
		ctx := logutil.WithRequestID(r.Context(), requestID)

		// Add to response header for client correlation
		w.Header().Set(RequestIDHeader, requestID)

		// Continue with request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
