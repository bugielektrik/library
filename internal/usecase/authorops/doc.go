// Package authorops implements use cases for author management operations.
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
//   - author.Repository: For author persistence
//   - author.Service: For business rule validation (if applicable)
//
// Example usage:
//
//	createUC := authorops.NewCreateAuthorUseCase(repo)
//	response, err := createUC.Execute(ctx, authorops.CreateAuthorRequest{
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
//   - Package name uses "ops" suffix to avoid conflict with domain author package
//   - Simple CRUD operations without complex business logic
//   - Validation handled by domain layer
package authorops
