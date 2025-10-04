// Package usecase contains application-specific business rules and use case orchestration.
//
// This package implements the use case layer of clean architecture, coordinating
// domain entities and services to fulfill specific application requirements.
//
// Responsibilities:
//   - Orchestrate domain entities and services
//   - Define transaction boundaries
//   - Transform between DTOs and domain models
//   - Implement application-specific workflows
//
// Design principles:
//   - One use case per business operation
//   - Single Responsibility Principle
//   - Dependency injection via constructors
//   - Use cases depend only on domain interfaces
//
// Structure:
//   - Each use case is a separate file (e.g., create_book.go)
//   - Use cases have Execute() methods for invocation
//   - Request/Response DTOs defined per use case
//   - Dependencies injected through constructors
//
// Example use case structure:
//
//	type CreateBookUseCase struct {
//		repo    book.Repository
//		service *book.Service
//	}
//
//	func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
//		// 1. Validate input
//		// 2. Create domain entity
//		// 3. Persist via repository
//		// 4. Return response
//	}
package usecase
