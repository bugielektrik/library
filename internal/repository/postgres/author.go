package postgres

import (
	"context"
	"fmt"
	"library/internal/entity"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// AuthorRepository is a postgres implementation of the AuthorRepository interface
type AuthorRepository struct {
	db *sqlx.DB
}

// NewAuthorRepository creates a new instance of the AuthorRepository struct
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

// CreateRow creates a new row in the postgres database
func (s *AuthorRepository) CreateRow(data entity.Author) (id string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		INSERT INTO authors (full_name, pseudonym, specialty)
		VALUES ($1, $2, $3)
		RETURNING id`

	args := []any{data.FullName, data.Pseudonym, data.Specialty}

	err = s.db.QueryRowContext(ctx, query, args...).Scan(&id)

	return
}

// GetRowByID retrieves a row from the postgres database by ID
func (s *AuthorRepository) GetRowByID(id string) (dest entity.Author, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		WHERE id=$1`

	args := []any{id}

	err = s.db.GetContext(ctx, &dest, query, args...)

	return
}

// SelectRows retrieves all rows from the postgres database
func (s *AuthorRepository) SelectRows() (dest []entity.Author, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, full_name, pseudonym, specialty
		FROM authors
		ORDER BY id`

	err = s.db.SelectContext(ctx, &dest, query)

	return
}

// UpdateRow updates an existing row in the postgres database
func (s *AuthorRepository) UpdateRow(id string, data entity.Author) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

// DeleteRow deletes a row from the postgres database by ID
func (s *AuthorRepository) DeleteRow(id string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		DELETE 
		FROM authors
		WHERE id=$1`

	args := []any{id}

	_, err = s.db.ExecContext(ctx, query, args...)

	return
}
