package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"library-service/internal/infrastructure/store"
)

// HandleSQLError converts common SQL errors to domain errors.
// This centralizes error handling logic across all postgres repositories.
//
// Conversions:
//   - sql.ErrNoRows → store.ErrorNotFound
//   - nil → nil (passthrough)
//   - other errors → returned as-is
func HandleSQLError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// BaseRepository provides common CRUD operations that can be embedded in other repositories.
// It uses Go generics to work with any entity type that has an ID field.
//
// Usage example:
//
//	type BookRepository struct {
//	    BaseRepository[book.Book]
//	}
//
//	func NewBookRepository(db *sqlx.DB) *BookRepository {
//	    return &BookRepository{
//	        BaseRepository: NewBaseRepository[book.Book](db, "books"),
//	    }
//	}
//
// Note: Add and Update methods are intentionally simple/incomplete as they require
// entity-specific column mappings. Repositories should override these methods.
type BaseRepository[T any] struct {
	db        *sqlx.DB
	tableName string
}

// NewBaseRepository creates a new base repository instance.
//
// Parameters:
//   - db: Database connection
//   - tableName: Name of the database table
func NewBaseRepository[T any](db *sqlx.DB, tableName string) BaseRepository[T] {
	return BaseRepository[T]{
		db:        db,
		tableName: tableName,
	}
}

// GenerateID generates a new UUID for entity IDs.
// Useful for repositories that need to generate IDs before insertion.
func (r *BaseRepository[T]) GenerateID() string {
	return uuid.New().String()
}

// Get retrieves a single entity by ID.
//
// Returns ErrNotFound if the entity doesn't exist.
func (r *BaseRepository[T]) Get(ctx context.Context, id string) (T, error) {
	return GetByID[T](ctx, r.db, r.tableName, id)
}

// List retrieves all entities from the table.
//
// Results are ordered by the 'id' column by default.
// Use ListWithOrder for custom ordering.
func (r *BaseRepository[T]) List(ctx context.Context) ([]T, error) {
	return List[T](ctx, r.db, r.tableName, "id")
}

// ListWithOrder retrieves all entities with custom ordering.
//
// Example: ListWithOrder(ctx, "created_at DESC")
func (r *BaseRepository[T]) ListWithOrder(ctx context.Context, orderBy string) ([]T, error) {
	return List[T](ctx, r.db, r.tableName, orderBy)
}

// Delete removes an entity by ID.
//
// Returns ErrNotFound if the entity doesn't exist.
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	return DeleteByID(ctx, r.db, r.tableName, id)
}

// Exists checks if an entity with the given ID exists.
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	return ExistsByID(ctx, r.db, r.tableName, id)
}

// Count returns the total number of entities in the table.
func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error) {
	return CountAll(ctx, r.db, r.tableName)
}

// GetDB returns the underlying database connection.
// Useful for repositories that need to execute custom queries.
func (r *BaseRepository[T]) GetDB() *sqlx.DB {
	return r.db
}

// GetTableName returns the table name for this repository.
func (r *BaseRepository[T]) GetTableName() string {
	return r.tableName
}

// BatchGet retrieves multiple entities by their IDs.
//
// Returns a slice of entities in the same order as the input IDs.
// Missing entities will be omitted from the result.
func (r *BaseRepository[T]) BatchGet(ctx context.Context, ids []string) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}

	query := fmt.Sprintf(`
		SELECT * FROM %s
		WHERE id = ANY($1)
		ORDER BY id
	`, r.tableName)

	var entities []T
	err := r.db.SelectContext(ctx, &entities, query, ids)
	return entities, HandleSQLError(err)
}

// Transaction executes a function within a database transaction.
//
// If the function returns an error, the transaction is rolled back.
// Otherwise, it's committed.
//
// Example:
//
//	err := repo.Transaction(ctx, func(tx *sqlx.Tx) error {
//	    // Execute multiple operations
//	    return nil
//	})
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
