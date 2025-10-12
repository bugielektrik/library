// Package http provides HTTP handler for book management service.
//
// This package handles comprehensive book-related HTTP requests including:
//   - Create new book (POST /books)
//   - Get book by ID (GET /books/{id})
//   - List all books with filtering (GET /books)
//   - Update book details (PUT /books/{id})
//   - Delete book (soft delete) (DELETE /books/{id})
//   - List book authors (GET /books/{id}/authors)
//
// Handler Organization:
//   - handler.go: Handler struct, routes, and constructor
//   - crud.go: Create, Read, Update, Delete operations
//   - query.go: List and search operations
//   - dto.go: Request/response DTOs
//
// This organization follows the CQRS pattern with separate files for:
//   - Commands (CRUD): Operations that modify state
//   - Queries: Operations that only read state
//
// All operations require authentication (JWT middleware applied in router).
//
// Related Packages:
//   - Use Cases: internal/books/service/ (book business logic)
//   - Domain: internal/books/domain/book/ (book entity, service, and repository interface)
//   - DTOs: dto.go in this package (request/response types)
//   - Cache: internal/adapters/cache/ (book caching for GET operations)
//
// Example Usage:
//
//	bookHandler := http.NewBookHandler(useCases, validator)
//	router.Group(func(r chi.Router) {
//	    r.Use(authMiddleware.Authenticate)
//	    r.Mount("/books", bookHandler.Routes())
//	})
package http
