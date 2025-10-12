// Package httputil provides HTTP-related utility functions.
package httputil

const (
	// HTTP Status Code Thresholds
	StatusInternalError = 500 // Server errors (5xx)
	StatusClientError   = 400 // Client errors (4xx)
	StatusRedirect      = 300 // Redirects (3xx)
	StatusSuccess       = 200 // Success responses (2xx)
)

// IsServerError checks if status code indicates a server error (5xx).
func IsServerError(code int) bool {
	return code >= StatusInternalError && code < 600
}

// IsClientError checks if status code indicates a client error (4xx).
func IsClientError(code int) bool {
	return code >= StatusClientError && code < StatusInternalError
}

// IsSuccess checks if status code indicates success (2xx).
func IsSuccess(code int) bool {
	return code >= StatusSuccess && code < StatusRedirect
}

// IsRedirect checks if status code indicates a redirect (3xx).
func IsRedirect(code int) bool {
	return code >= StatusRedirect && code < StatusClientError
}

// IsError checks if status code indicates any error (4xx or 5xx).
func IsError(code int) bool {
	return code >= StatusClientError
}
