package postgres

import (
	"context"
	"fmt"
	"library/internal/entity"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// BookRepository is a postgres implementation of the BookRepository interface
type BookRepository struct {
	db *sqlx.DB
}

// NewBookRepository creates a new instance of the BookRepository struct
func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

// CreateRow creates a new row in the postgres database
func (s *BookRepository) CreateRow(data entity.Book) (id string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO books (name, genre, isbn, authors)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.Name, data.Genre, data.ISBN, data.Authors}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

// GetRowByID retrieves a row from the postgres database by ID
func (s *BookRepository) GetRowByID(id string) (dest entity.Book, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

// SelectRows retrieves all rows from the postgres database
func (s *BookRepository) SelectRows() (dest []entity.Book, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

// UpdateRow updates an existing row in the postgres database
func (s *BookRepository) UpdateRow(id string, data entity.Book) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE books SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = s.db.ExecContext(ctx, query, args...)
	}

	return
}

func (s *BookRepository) prepareArgs(data entity.Book) (sets []string, args []any) {
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
		args = append(args, data.Authors)
		sets = append(sets, fmt.Sprintf("authors=$%d", len(args)))
	}

	return
}

// DeleteRow deletes a row from the postgres database by ID
func (s *BookRepository) DeleteRow(id string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		DELETE 
		FROM books
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
