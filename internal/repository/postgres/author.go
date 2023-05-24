package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library/internal/entity"
)

type AuthorRepository struct {
	db *sqlx.DB
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

func (s *AuthorRepository) SelectRows(ctx context.Context) (dest []entity.Author, err error) {
	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

func (s *AuthorRepository) CreateRow(ctx context.Context, data entity.Author) (id string, err error) {
	query := `
		INSERT INTO authors (full_name, pseudonym, specialty)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.FullName, data.Pseudonym, data.Specialty}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

func (s *AuthorRepository) GetRow(ctx context.Context, id string) (dest entity.Author, err error) {
	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

func (s *AuthorRepository) UpdateRow(ctx context.Context, id string, data entity.Author) (err error) {
	sets, args := s.prepareArgs(data)
	if len(args) > 0 {

		args = append(args, id)
		sets = append(sets, "updated_at=CURRENT_TIMESTAMP")

		query := fmt.Sprintf("UPDATE authors SET %s WHERE id=$%d", strings.Join(sets, ", "), len(args))
		_, err = s.db.ExecContext(ctx, query, args...)
	}

	return
}

func (s *AuthorRepository) prepareArgs(data entity.Author) (sets []string, args []any) {
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

func (s *AuthorRepository) DeleteRow(ctx context.Context, id string) (err error) {
	query := `
		DELETE 
		FROM authors
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
