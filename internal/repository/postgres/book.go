package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library/internal/domain/book"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (s *BookRepository) Select(ctx context.Context) (dest []book.Entity, err error) {
	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

func (s *BookRepository) Create(ctx context.Context, data book.Entity) (id string, err error) {
	fmt.Println(data.Authors.String())
	query := `
		INSERT INTO books (name, genre, isbn, authors)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	args := []any{data.Name, data.Genre, data.ISBN, data.Authors.String()}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

func (s *BookRepository) Get(ctx context.Context, id string) (dest book.Entity, err error) {
	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *BookRepository) Update(ctx context.Context, id string, data book.Entity) (err error) {
	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE books SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = s.db.ExecContext(ctx, query, args...)
	}

	return
}

func (s *BookRepository) prepareArgs(data book.Entity) (sets []string, args []any) {
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
		args = append(args, data.Authors.String())
		sets = append(sets, fmt.Sprintf("authors=$%d", len(args)))
	}

	return
}

func (s *BookRepository) Delete(ctx context.Context, id string) (err error) {
	query := `
		DELETE 
		FROM books
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
