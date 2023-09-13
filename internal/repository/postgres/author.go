package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/author"
	"library-service/pkg/store"
)

type AuthorRepository struct {
	db *sqlx.DB
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

func (r *AuthorRepository) List(ctx context.Context) (dest []author.Entity, err error) {
	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		ORDER BY id`

	err = r.db.SelectContext(ctx, &dest, query)

	return
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (id string, err error) {
	query := `
		INSERT INTO authors (full_name, pseudonym, specialty)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.FullName, data.Pseudonym, data.Specialty}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (dest author.Entity, err error) {
	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		WHERE id=$1`

	args := []any{id}

	if err = r.db.GetContext(ctx, &dest, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) (err error) {
	sets, args := r.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
		query := fmt.Sprintf("UPDATE authors SET %r WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))

		if err = r.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = store.ErrorNotFound
			}
		}
	}

	return
}

func (r *AuthorRepository) prepareArgs(data author.Entity) (sets []string, args []any) {
	if data.Pseudonym != nil {
		args = append(args, data.Pseudonym)
		sets = append(sets, fmt.Sprintf("pseudonym=$%d", len(args)))
	}

	if data.FullName != nil {
		args = append(args, data.FullName)
		sets = append(sets, fmt.Sprintf("full_name=$%d", len(args)))
	}

	if data.Specialty != nil {
		args = append(args, data.Specialty)
		sets = append(sets, fmt.Sprintf("specialty=$%d", len(args)))
	}

	return
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) (err error) {
	query := `
		DELETE FROM authors
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
