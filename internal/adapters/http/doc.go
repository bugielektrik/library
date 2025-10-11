// Package http provides HTTP REST API handlers and routing for the library service.
//
// This package implements the inbound HTTP adapter in clean architecture,
// translating HTTP requests into use case calls and formatting responses.
// It uses Chi router for request routing and middleware.
//
// Package structure:
//   - handlers/: HTTP request handlers organized by domain entity
//   - dto/: Data Transfer Objects for request/response serialization
//   - middleware/: HTTP middleware (auth, logging, CORS, rate limiting)
//   - router.go: Main router configuration and route registration
//   - http.go: HTTP server initialization and configuration
//
// Handlers:
//   - BookHandler: CRUD operations for books
//   - MemberHandler: Member profile management
//   - AuthHandler: Authentication endpoints (login, register, refresh)
//   - PaymentHandler: Payment processing and callbacks
//   - ReservationHandler: Book reservation management
//
// DTOs:
//   - Request DTOs: Validate and deserialize incoming JSON
//   - Response DTOs: Serialize domain entities to JSON
//   - Validation tags: go-playground/validator for input validation
//
// Middleware:
//   - JWT authentication: Validates bearer tokens, injects member context
//   - Request logging: Structured logging with request ID, duration
//   - Error handling: Translates domain errors to HTTP status codes
//   - CORS: Cross-origin resource sharing configuration
//   - Rate limiting: Request throttling per IP/member
//
// Example handler implementation:
//
//	type BookHandler struct {
//	    createBookUC *bookops.CreateBookUseCase
//	}
//
//	func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
//	    var req dto.CreateBookRequest
//	    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//	        h.respondError(w, r, errors.ErrInvalidRequest)
//	        return
//	    }
//	    // Call use case, return response
//	}
//
// API documentation:
//   - Swagger/OpenAPI annotations on all handler methods
//   - Documentation generated via swaggo/swag
//   - Available at: /swagger/index.html
//
// Error handling:
//   - Domain errors mapped to HTTP status codes
//   - Structured error responses with error codes and details
//   - Sensitive error details excluded from responses
//
// Authentication:
//   - JWT bearer tokens in Authorization header
//   - Protected endpoints require valid token
//   - Public endpoints: /auth/login, /auth/register, /health
package http
