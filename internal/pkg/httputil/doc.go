// Package httputil provides HTTP-related utility functions and constants.
//
// This package contains helpers for HTTP operations commonly needed across
// the application, particularly when working with HTTP status codes and responses.
//
// Utilities:
//   - HTTP status code constants and thresholds
//   - Status code validation functions (IsServerError, IsClientError, etc.)
//
// Example usage:
//
//	import "library-service/internal/infrastructure/pkg/httputil"
//
//	// Check status code type
//	if httputil.IsServerError(statusCode) {
//	    logger.Error("server error occurred")
//	} else if httputil.IsClientError(statusCode) {
//	    logger.Warn("client error occurred")
//	}
//
//	// Use constants for clarity
//	if statusCode >= httputil.StatusInternalError {
//	    // Handle server error
//	}
package httputil
