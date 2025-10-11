// Package repository provides persistence layer implementations for domain entities.
//
// This package implements the outbound repository adapter in clean architecture,
// providing concrete implementations of domain repository interfaces for various
// storage backends.
//
// Implementations:
//   - postgres/: PostgreSQL repository implementations (primary database)
//   - mongo/: MongoDB repository implementations (alternative NoSQL backend)
//   - memory/: In-memory repository implementations (testing, development)
//   - mocks/: Generated mock repositories for use case testing
//
// PostgreSQL repositories:
//   - BookRepository: Books with ISBN indexing
//   - MemberRepository: Members with email uniqueness constraint
//   - AuthorRepository: Authors with name indexing
//   - ReservationRepository: Reservations with composite indexes
//   - PaymentRepository: Payment records with transaction safety
//
// Design principles:
//   - Each repository implements domain repository interface
//   - Database-specific SQL in repository implementation
//   - Transaction management handled at repository level
//   - Query optimization via prepared statements
//   - Connection pooling configured in infrastructure layer
//
// Example PostgreSQL repository:
//
//	type PostgresBookRepository struct {
//	    db *sqlx.DB
//	}
//
//	func (r *PostgresBookRepository) Add(ctx context.Context, book book.Book) (string, error) {
//	    query := `INSERT INTO books (id, name, genre, isbn) VALUES ($1, $2, $3, $4)`
//	    _, err := r.db.ExecContext(ctx, query, book.ID, book.Name, book.Genre, book.ISBN)
//	    return book.ID, err
//	}
//
// Query patterns:
//   - Context-aware queries for timeout and cancellation
//   - Parameterized queries to prevent SQL injection
//   - Named parameters with sqlx for readability
//   - Batch operations where applicable
//
// Error handling:
//   - Database constraint violations mapped to domain errors
//   - Connection errors wrapped with context
//   - Deadlocks and conflicts handled with retries
//
// Testing:
//   - Integration tests use real PostgreSQL (docker-compose)
//   - Unit tests use mocks or memory implementations
//   - Test fixtures provided in test/fixtures/
//
// Performance considerations:
//   - Indexes on frequently queried columns (ID, ISBN, email)
//   - Composite indexes for multi-column queries
//   - Connection pooling (max 25 connections by default)
//   - Query timeout: 5 seconds default, configurable
package repository
