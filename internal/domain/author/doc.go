// Package author provides the core business logic and entities for author management.
//
// This package implements the domain layer of Clean Architecture, containing:
//   - Author entity with personal information
//   - Repository interface for data persistence
//   - Cache interface for performance optimization
//   - DTOs for data transfer
//
// Business Rules:
//   - Author must have a full name
//   - Pseudonym is optional
//   - Specialty/genre is optional but recommended
//   - Authors can write multiple books
//
// Example Usage:
//
//	// Create author entity
//	author := author.Entity{
//	    ID:        "123",
//	    FullName:  "J.K. Rowling",
//	    Pseudonym: "Robert Galbraith",
//	    Specialty: "Fantasy, Crime Fiction",
//	}
//
//	// Use repository to persist
//	err := authorRepo.Create(ctx, author)
//
//	// Use cache for fast lookups
//	cached, err := authorCache.Get(ctx, "123")
//
// The author domain follows the same Clean Architecture principles as other
// domains, maintaining independence from external frameworks and databases.
package author
