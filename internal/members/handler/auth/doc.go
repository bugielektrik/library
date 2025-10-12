// Package auth provides HTTP handler for authentication and authorization service.
//
// This package implements JWT-based authentication including:
//   - User registration with email/password validation (POST /auth/register)
//   - Login with credential verification (POST /auth/login)
//   - JWT token refresh for session management (POST /auth/refresh)
//   - Current user profile retrieval (GET /auth/me)
//
// All endpoints follow the handler patterns defined in the parent handler package.
// Protected endpoints require authentication middleware to validate JWT tokens.
//
// Handler Organization:
//   - handler.go: Handler struct, routes, constructor, and all endpoint implementations
//
// Authentication Flow:
//  1. Register: Create member → Hash password → Generate JWT tokens
//  2. Login: Verify credentials → Generate JWT tokens
//  3. Refresh: Validate refresh token → Issue new access token
//  4. Protected requests: JWT middleware validates token → Extract member ID → Pass to handler
//
// Related Packages:
//   - Use Cases: internal/usecase/authops/ (business logic)
//   - Domain: internal/domain/member/ (member entity)
//   - DTOs: internal/adapters/http/dto/member.go (request/response types)
//   - Middleware: internal/adapters/http/middleware/auth.go (JWT validation)
//   - Infrastructure: internal/infrastructure/auth/ (JWT + password service)
//
// Example Usage:
//
//	authHandler := auth.NewAuthHandler(useCases, validator)
//	router.Route("/auth", func(r chi.Router) {
//	    r.Post("/register", authHandler.Register)
//	    r.Post("/login", authHandler.Login)
//	    r.Post("/refresh", authHandler.RefreshToken)
//	    r.With(authMiddleware).Get("/me", authHandler.GetCurrentMember)
//	})
package auth
