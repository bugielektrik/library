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
	"library-service/pkg/store"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) List(ctx context.Context) (dest []book.Entity, err error) {
	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		ORDER BY id`

	err = r.db.SelectContext(ctx, &dest, query)

	return
}

func (r *BookRepository) Add(ctx context.Context, data book.Entity) (id string, err error) {
	query := `
		INSERT INTO books (name, genre, isbn, authors)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	args := []any{data.Name, data.Genre, data.ISBN, pq.Array(data.Authors)}

	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *BookRepository) Get(ctx context.Context, id string) (dest book.Entity, err error) {
	query := `
		SELECT id, name, genre, isbn, authors
		FROM books
		WHERE id=$1`

	args := []any{id}

	if err = r.db.GetContext(ctx, &dest, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *BookRepository) Update(ctx context.Context, id string, data book.Entity) (err error) {
	sets, args := r.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
		query := fmt.Sprintf("UPDATE books SET %r WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))

		if err = r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = store.ErrorNotFound
			}
		}
	}

	return
}

func (r *BookRepository) prepareArgs(data book.Entity) (sets []string, args []any) {
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

	return
}

func (r *BookRepository) Delete(ctx context.Context, id string) (err error) {
	query := `
		DELETE FROM books
		WHERE id=$1
		RETURNING id`

	args := []any{id}

	if err = r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrorNotFound
		}
	}

	return
}
