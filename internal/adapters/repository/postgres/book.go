package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"library-service/internal/domain/book"
	"library-service/internal/infrastructure/store"
)

// BookRepository handles CRUD operations for books in a PostgreSQL store.
type BookRepository struct {
	db *sqlx.DB
}

// NewBookRepository creates a new BookRepository.
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

// List retrieves all books from the store.
func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) {
	query := `SELECT id, name, genre, isbn, authors FROM books ORDER BY id`
	var books []book.Book
	err := r.db.SelectContext(ctx, &books, query)
	return books, err
}

// Add inserts a new book into the store.
func (r *BookRepository) Add(ctx context.Context, data book.Book) (string, error) {
	query := `INSERT INTO books (name, genre, isbn, authors) VALUES ($1, $2, $3, $4) RETURNING id`
	args := []interface{}{data.Name, data.Genre, data.ISBN, pq.Array(data.Authors)}
	var id string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", store.ErrorNotFound
	}
	return id, err
}

// Get retrieves a book by ID from the store.
func (r *BookRepository) Get(ctx context.Context, id string) (book.Book, error) {
	query := `SELECT id, name, genre, isbn, authors FROM books WHERE id=$1`
	var book book.Book
	err := r.db.GetContext(ctx, &book, query, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return book, store.ErrorNotFound
	}
	return book, err
}

// Update modifies an existing book in the store.
func (r *BookRepository) Update(ctx context.Context, id string, data book.Book) error {
	sets, args := r.prepareArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE books SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// Delete removes a book by ID from the store.
func (r *BookRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM books WHERE id=$1 RETURNING id`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

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
