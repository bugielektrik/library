// Package reservation implements the reservation domain model.
//
// This package contains the core business logic and entities for book reservations,
// following Domain-Driven Design principles.
//
// Key Concepts:
//   - Reservation: Represents a member's request to reserve a book when it becomes available
//   - Status: Tracks the reservation lifecycle (pending, fulfilled, cancelled, expired)
//   - Service: Encapsulates business rules for reservation management
//
// Business Rules:
//   - Members can reserve books that are currently borrowed
//   - A member cannot reserve the same book twice if they have an active reservation
//   - A member cannot reserve a book they currently have borrowed
//   - Reservations expire after 7 days if not fulfilled
//   - When a book becomes available, the oldest pending reservation is fulfilled first
//   - Only pending or fulfilled reservations can be cancelled
package reservation
