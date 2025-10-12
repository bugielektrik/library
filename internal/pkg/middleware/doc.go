// Package middleware provides HTTP middleware for request processing.
//
// This package contains middleware components that intercept HTTP requests
// for cross-cutting concerns such as:
//   - Authentication and authorization (JWT validation)
//   - Request validation
//   - Logging and request tracing
//   - Error handling
//   - CORS handling
//   - Rate limiting
//
// Middleware functions follow the chi.Middleware pattern and can be
// composed in the router configuration.
package middleware
