// Package author implements use cases for author management service.
//
// This package orchestrates author-related workflows including author creation,
// retrieval, updates, and association with books. Authors represent the creators
// of books in the library system.
//
// Use cases implemented:
//   - CreateAuthorUseCase: Creates a new author with validation
//   - GetAuthorUseCase: Retrieves author details by ID
//   - ListAuthorsUseCase: Returns all authors with optional filtering
//   - UpdateAuthorUseCase: Updates author information
//   - DeleteAuthorUseCase: Removes an author (with book association checks)
//
// Dependencies:
//   - authordomain.Repository: For author persistence
//   - authordomain.Service: For business rule validation (if applicable)
//
// Example usage:
//
//	createUC := author.NewCreateAuthorUseCase(repo)
//	response, err := createUC.Execute(ctx, author.CreateAuthorRequest{
//	    Name:        "Robert C. Martin",
//	    Biography:   "Software engineer and author",
//	    Country:     "USA",
//	})
//
// Business rules:
//   - Author name must be unique
//   - Cannot delete author if associated with books
//   - Biography and country are optional fields
//
// Architecture:
//   - Part of books bounded context operations
//   - Simple CRUD operations without complex business logic
//   - Validation handled by domain layer
package author
