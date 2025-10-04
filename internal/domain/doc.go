// Package domain contains the core business logic of the Library Management System.
//
// This is the innermost layer of Clean Architecture, containing:
//   - Business entities (book, member, author)
//   - Domain services with business rules
//   - Repository and cache interfaces
//   - Value objects and domain DTOs
//
// Dependency Rule:
// The domain layer has NO dependencies on outer layers. It defines interfaces
// that outer layers must implement (Dependency Inversion Principle).
//
// Domain Packages:
//   - book: Book entities, ISBN validation, business rules
//   - member: Member entities, subscription logic, pricing rules
//   - author: Author entities and relationships
//
// Design Principles:
//   - Rich Domain Model: Business logic lives here, not in use cases
//   - Domain Services: Complex logic that doesn't belong to a single entity
//   - Repository Pattern: Interface defined here, implemented in adapters
//   - Testability: Pure business logic, no external dependencies
//
// Example:
//
//	// Domain service validates business rules
//	bookService := book.NewService()
//	if err := bookService.ValidateISBN(isbn); err != nil {
//	    return err
//	}
//
//	// Use case orchestrates domain logic
//	book := book.Entity{...}
//	if err := bookService.ValidateBook(book); err != nil {
//	    return err
//	}
//	return bookRepo.Create(ctx, book)
package domain
