// Package postgres provides PostgreSQL repository implementations.
//
// This package contains concrete implementations of repository interfaces
// defined in the domain layer, using PostgreSQL as the data store.
//
// Key components:
//   - BaseRepository: Generic CRUD operations using Go generics
//   - Entity-specific repositories: Book, Member, Author, Reservation, etc.
//   - SQL error handling and mapping
//   - Database transaction support
//
// All repositories use sqlx for database operations and follow the
// repository pattern to keep domain logic independent of database technology.
package postgres
