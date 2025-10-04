// Package domain contains the core business logic and entities for the library management system.
//
// This package follows Domain-Driven Design (DDD) principles and contains:
//   - Business entities (Book, Member, Author)
//   - Domain services encapsulating business rules
//   - Repository and cache interfaces (implementation-agnostic)
//   - Value objects and domain events
//
// The domain layer is the innermost layer of the clean architecture and has zero dependencies
// on external frameworks or infrastructure. All business rules and validation logic reside here.
//
// Subpackages:
//   - book: Book management with ISBN validation and lifecycle rules
//   - member: Member and subscription management with pricing logic
//   - author: Author information and metadata management
//
// Design principles:
//   - Domain entities are independent of persistence technology
//   - Repository interfaces defined here, implemented in adapters
//   - Business rules enforced through domain services
//   - Immutable value objects where appropriate
package domain
