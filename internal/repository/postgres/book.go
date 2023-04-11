package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"library/internal/entity"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}
func (s *BookRepository) CreateRow(data entity.Book) (dest string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO books (name,genre,codeisbn)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.Name, data.Genre, data.ISBN}
	err = s.db.QueryRowContext(ctx, query, args...).Scan(&dest)

	return
}

func (s *BookRepository) GetRowByID(id string) (dest entity.Book, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	query := `
		SELECT id, name, genre ,codeisbn
		FROM books
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *BookRepository) SelectRows() (dest []entity.Book, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT  id, name, genre, codeisbn
		FROM books
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

func (s *BookRepository) UpdateRow(data entity.Book) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, data.ID)
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
		sets = append(sets, fmt.Sprintf("codeISBN=$%d", len(args)))
	}
	return
}

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
