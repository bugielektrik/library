// Package author provides domain entities and business logic for author management.
//
// This package implements author-related domain logic including:
//   - Author entity with metadata (full name, pseudonym, specialty)
//   - Author validation rules
//   - Repository and cache interfaces for author persistence
//
// The author entity represents writers and creators of books in the library system,
// supporting both real names and pseudonyms.
//
// Example usage:
//
//	// Create new author
//	newAuthor := author.New(author.Request{
//	    FullName:  "Robert C. Martin",
//	    Pseudonym: "Uncle Bob",
//	    Specialty: "Software Architecture",
//	})
//
//	// Add to repository
//	authorID, err := repo.Add(ctx, newAuthor)
//
// Domain Rules:
//   - Author must have a valid full name (non-empty)
//   - Pseudonym is optional
//   - Specialty is optional but recommended for categorization
//   - Authors can be associated with multiple books
package author
