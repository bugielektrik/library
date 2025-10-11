// Package member provides HTTP handlers for member management operations.
//
// This package handles member-related HTTP requests including:
//   - List all members with pagination (GET /members)
//   - Get member profile by ID (GET /members/{id})
//   - Subscribe member to premium tier (POST /members/{id}/subscribe)
//
// All endpoints require authentication (JWT middleware applied in router).
//
// Handler Organization:
//   - handler.go: Handler struct, routes, constructor, and all endpoint implementations
//
// Related Packages:
//   - Use Cases: internal/usecase/memberops/ (member operations)
//   - Use Cases: internal/usecase/subops/ (subscription logic)
//   - Domain: internal/domain/member/ (member entity and service)
//   - DTOs: internal/adapters/http/dto/member.go (request/response types)
//
// Example Usage:
//
//	memberHandler := member.NewMemberHandler(useCases)
//	router.Group(func(r chi.Router) {
//	    r.Use(authMiddleware.Authenticate)
//	    r.Mount("/members", memberHandler.Routes())
//	})
package member
