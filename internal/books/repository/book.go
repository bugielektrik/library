package repository

import (
	"context"
	"fmt"
	postgres2 "library-service/internal/pkg/repository/postgres"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/books/domain/book"
)

// BookRepository handles CRUD operations for books in a PostgreSQL store.
type BookRepository struct {
	postgres2.BaseRepository[book.Book]
}

// Compile-time check that BookRepository implements book.Repository
var _ book.Repository = (*BookRepository)(nil)

// NewBookRepository creates a new BookRepository.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		BaseRepository: postgres2.NewBaseRepository[book.Book](db, "books"),
	}
}

// List is inherited from BaseRepository

// Add inserts a new book into the store.
func (r *BookRepository) Add(ctx context.Context, data book.Book) (string, error) {
	query := `INSERT INTO books (name, genre, isbn, authors) VALUES ($1, $2, $3, $4) RETURNING id`
	args := []interface{}{data.Name, data.Genre, data.ISBN, pq.Array(data.Authors)}
	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return id, postgres2.HandleSQLError(err)
}

// Get is inherited from BaseRepository

// Update modifies an existing book in the store.
func (r *BookRepository) Update(ctx context.Context, id string, data book.Book) error {
	sets, args := postgres2.PrepareUpdateArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE books SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return postgres2.HandleSQLError(err)
}

// Delete is inherited from BaseRepository
