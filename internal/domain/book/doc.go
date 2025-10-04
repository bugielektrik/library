// Package book provides the core business logic and entities for book management.
//
// This package implements the domain layer of Clean Architecture, containing:
//   - Book entity with business rules and validation
//   - Domain service for ISBN validation and business logic
//   - Repository interface for data persistence
//   - Cache interface for performance optimization
//
// Business Rules:
//   - ISBN must be valid (ISBN-10 or ISBN-13 with checksum)
//   - Book must have at least one author
//   - Book name is required
//   - Books can only be deleted if not borrowed by any member
//
// Example Usage:
//
//	// Create domain service
//	service := book.NewService()
//
//	// Validate ISBN
//	if err := service.ValidateISBN("978-0-306-40615-7"); err != nil {
//	    return err
//	}
//
//	// Validate complete book
//	bookEntity := book.Entity{
//	    ID:      "123",
//	    Name:    "Clean Architecture",
//	    ISBN:    "978-0-134-49416-6",
//	    Authors: []string{"Robert C. Martin"},
//	}
//	if err := service.ValidateBook(bookEntity); err != nil {
//	    return err
//	}
//
// The book domain is independent of external frameworks and can be tested
// in isolation without database or HTTP dependencies.
package book
