package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/domain/book"
)

// BookRepository handles CRUD operations for books in a PostgreSQL store.
type BookRepository struct {
	BaseRepository[book.Book]
}

// NewBookRepository creates a new BookRepository.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		BaseRepository: NewBaseRepository[book.Book](db, "books"),
	}
}

// List is inherited from BaseRepository

// Add inserts a new book into the store.
func (r *BookRepository) Add(ctx context.Context, data book.Book) (string, error) {
	query := `INSERT INTO books (name, genre, isbn, authors) VALUES ($1, $2, $3, $4) RETURNING id`
	args := []interface{}{data.Name, data.Genre, data.ISBN, pq.Array(data.Authors)}
	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return id, HandleSQLError(err)
}

// Get is inherited from BaseRepository

// Update modifies an existing book in the store.
func (r *BookRepository) Update(ctx context.Context, id string, data book.Book) error {
	sets, args := r.prepareArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE books SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return HandleSQLError(err)
}

// Delete is inherited from BaseRepository

// prepareArgs prepares the update arguments for the SQL query.
func (r *BookRepository) prepareArgs(data book.Book) ([]string, []interface{}) {
	var sets []string
	var args []interface{}

	if data.Name != nil {
		args = append(args, data.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}
	if data.Genre != nil {
		args = append(args, data.Genre)
		sets = append(sets, fmt.Sprintf("genre=$%d", len(args)))
	}
	if data.ISBN != nil {
		args = append(args, data.ISBN)
		sets = append(sets, fmt.Sprintf("isbn=$%d", len(args)))
	}
	if len(data.Authors) > 0 {
		args = append(args, pq.Array(data.Authors))
		sets = append(sets, fmt.Sprintf("authors=$%d", len(args)))
	}

	return sets, args
}
