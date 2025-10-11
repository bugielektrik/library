// Package operations implements use cases for book management operations.
//
// This package orchestrates book-related business workflows by coordinating
// domain entities, services, and repositories. Each use case represents a
// specific book operation (create, read, update, delete, list).
//
// Use cases implemented:
//   - CreateBookUseCase: Creates a new book with validation and caching
//   - GetBookUseCase: Retrieves a single book by ID or ISBN with cache lookup
//   - ListBooksUseCase: Returns all books with optional filtering
//   - UpdateBookUseCase: Updates book information with validation
//   - DeleteBookUseCase: Removes a book after business rule checks
//   - ListBookAuthorsUseCase: Retrieves authors for a specific book
//
// Dependencies:
//   - book.Repository: For book persistence
//   - book.Cache: For performance optimization
//   - book.Service: For business rule validation
//
// Example usage:
//
//	createUC := operations.NewCreateBookUseCase(repo, cache, service)
//	response, err := createUC.Execute(ctx, operations.CreateBookRequest{
//	    Name:    "Clean Code",
//	    Genre:   "Programming",
//	    ISBN:    "978-0132350884",
//	    Authors: []string{"author-uuid"},
//	})
//
// Architecture:
//   - Package name "operations" to represent book-specific operations within the books bounded context
//   - Each use case follows the pattern: struct with dependencies, Execute() method
//   - Request/Response types defined per use case for type safety
//   - All use cases accept context.Context as first parameter
//
// Error handling:
//   - Validation errors: Return domain errors (e.g., errors.ErrInvalidISBN)
//   - Repository errors: Wrapped with context (e.g., errors.Database("database operation", err))
//   - Cache errors: Logged but not propagated (cache failures are non-critical)
package operations
