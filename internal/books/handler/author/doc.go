// Package author provides HTTP handler for author management service.
//
// This package handles author-related HTTP requests including:
//   - List all authors with pagination support (GET /authors)
//
// Handler Organization:
//   - handler.go: Handler struct, routes, constructor, and endpoint implementations
//
// Related Packages:
//   - Use Cases: internal/usecase/authorops/ (author business logic)
//   - Domain: internal/books/domain/author/ (author entity and repository interface)
//   - DTOs: internal/adapters/http/dto/author.go (request/response types)
//   - Cache: internal/adapters/cache/ (author caching)
//
// Example Usage:
//
//	authorHandler := author.NewAuthorHandler(useCases)
//	router.Mount("/authors", authorHandler.Routes())
package author
