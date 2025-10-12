package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// GenericRepository provides generic CRUD operations for entities.
// This eliminates repetitive code across all repositories.
//
// Type parameter T represents the entity type (e.g., book.Book, author.Author).
//
// Usage example:
//
//	type BookRepository struct {
//	    db *sqlx.DB
//	}
//
//	func (r *BookRepository) Get(ctx context.Context, id string) (book.Book, error) {
//	    return GetByID[book.Book](ctx, r.db, "books", id)
//	}
type GenericRepository[T any] struct {
	db        *sqlx.DB
	tableName string
}

// NewGenericRepository creates a new generic repository for the specified table.
func NewGenericRepository[T any](db *sqlx.DB, tableName string) *GenericRepository[T] {
	return &GenericRepository[T]{
		db:        db,
		tableName: tableName,
	}
}

// GetByID retrieves a single entity by ID from the database.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table to query
//   - id: ID of the entity to retrieve
//
// Returns the entity or an error. sql.ErrNoRows is converted to store.ErrorNotFound.
//
// Usage:
//
//	author, err := GetByID[author.Author](ctx, db, "authors", id)
func GetByID[T any](ctx context.Context, db *sqlx.DB, tableName string, id string) (T, error) {
	var entity T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=$1", tableName)
	err := db.GetContext(ctx, &entity, query, id)
	return entity, HandleSQLError(err)
}

// GetByIDWithColumns retrieves a single entity by ID with specific columns.
//
// This is useful when you need to select specific columns instead of SELECT *.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table to query
//   - columns: Comma-separated list of columns (e.g., "id, name, email")
//   - id: ID of the entity to retrieve
//
// Usage:
//
//	book, err := GetByIDWithColumns[book.Book](ctx, db, "books", "id, name, genre, isbn, authors", id)
func GetByIDWithColumns[T any](ctx context.Context, db *sqlx.DB, tableName string, columns string, id string) (T, error) {
	var entity T
	query := fmt.Sprintf("SELECT %s FROM %s WHERE id=$1", columns, tableName)
	err := db.GetContext(ctx, &entity, query, id)
	return entity, HandleSQLError(err)
}

// List retrieves all entities from a table.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table to query
//   - orderBy: Optional ORDER BY clause (e.g., "id", "created_at DESC"). If empty, defaults to "id"
//
// Returns a slice of entities or an error.
//
// Usage:
//
//	authors, err := List[author.Author](ctx, db, "authors", "id")
func List[T any](ctx context.Context, db *sqlx.DB, tableName string, orderBy string) ([]T, error) {
	var entities []T
	if orderBy == "" {
		orderBy = "id"
	}
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY %s", tableName, orderBy)
	err := db.SelectContext(ctx, &entities, query)
	return entities, err
}

// ListWithColumns retrieves all entities from a table with specific columns.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table to query
//   - columns: Comma-separated list of columns to select
//   - orderBy: Optional ORDER BY clause. If empty, defaults to "id"
//
// Usage:
//
//	books, err := ListWithColumns[book.Book](ctx, db, "books", "id, name, genre, isbn, authors", "id")
func ListWithColumns[T any](ctx context.Context, db *sqlx.DB, tableName string, columns string, orderBy string) ([]T, error) {
	var entities []T
	if orderBy == "" {
		orderBy = "id"
	}
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s", columns, tableName, orderBy)
	err := db.SelectContext(ctx, &entities, query)
	return entities, err
}

// DeleteByID deletes an entity by ID from the database.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table
//   - id: ID of the entity to delete
//
// Returns an error if the entity doesn't exist or the delete fails.
//
// Usage:
//
//	err := DeleteByID(ctx, db, "authors", id)
func DeleteByID(ctx context.Context, db *sqlx.DB, tableName string, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1 RETURNING id", tableName)
	err := db.QueryRowContext(ctx, query, id).Scan(&id)
	return HandleSQLError(err)
}

// ExistsByID checks if an entity exists in the database.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table
//   - id: ID of the entity to check
//
// Returns true if the entity exists, false otherwise.
//
// Usage:
//
//	exists, err := ExistsByID(ctx, db, "books", bookID)
func ExistsByID(ctx context.Context, db *sqlx.DB, tableName string, id string) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1)", tableName)
	var exists bool
	err := db.GetContext(ctx, &exists, query, id)
	return exists, err
}

// CountAll returns the total number of rows in a table.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - db: Database connection
//   - tableName: Name of the table
//
// Usage:
//
//	count, err := CountAll(ctx, db, "members")
func CountAll(ctx context.Context, db *sqlx.DB, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int64
	err := db.GetContext(ctx, &count, query)
	return count, err
}
