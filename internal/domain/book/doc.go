// Package book provides domain entities and business logic for book management.
//
// This package implements book-related domain logic including:
//   - Book entity with ISBN validation
//   - Book lifecycle rules (creation, update, deletion)
//   - ISBN format validation (ISBN-10 and ISBN-13)
//   - Repository and cache interfaces for book persistence
//
// The book entity represents physical or digital books in the library system,
// with strict validation rules enforced through domain services.
//
// Example usage:
//
//	service := book.NewService()
//
//	// Validate ISBN before creating book
//	if err := service.ValidateISBN("978-0134190440"); err != nil {
//	    return err
//	}
//
//	// Create new book
//	newBook := book.New(book.Request{
//	    Name:    "The Go Programming Language",
//	    Genre:   "Programming",
//	    ISBN:    "978-0134190440",
//	    Authors: []string{"author-id-1", "author-id-2"},
//	})
//
// ISBN Validation:
//   - Supports both ISBN-10 (10 digits) and ISBN-13 (13 digits with 978/979 prefix)
//   - Validates checksum for both formats
//   - Removes hyphens and spaces before validation
//
// Domain Rules:
//   - Book must have a valid name (non-empty)
//   - ISBN must be valid if provided
//   - Genre is optional but recommended
//   - Books can have multiple authors
package book
