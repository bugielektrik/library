package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/author"
	"library-service/internal/infrastructure/database"
)

// AuthorRepository handles CRUD operations for authors in a PostgreSQL database.
type AuthorRepository struct {
	db *sqlx.DB
}

// NewAuthorRepository creates a new AuthorRepository.
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// List retrieves all authors from the database.
func (r *AuthorRepository) List(ctx context.Context) ([]author.Entity, error) {
	query := `SELECT id, full_name, pseudonym, specialty FROM authors ORDER BY id`
	var authors []author.Entity
	err := r.db.SelectContext(ctx, &authors, query)
	return authors, err
}

// Add inserts a new author into the database.
func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	query := `INSERT INTO authors (full_name, pseudonym, specialty) VALUES ($1, $2, $3) RETURNING id`
	args := []interface{}{data.FullName, data.Pseudonym, data.Specialty}
	var id string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", store.ErrorNotFound
	}
	return id, err
}

// Get retrieves an author by ID from the database.
func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {
	query := `SELECT id, full_name, pseudonym, specialty FROM authors WHERE id=$1`
	var author author.Entity
	err := r.db.GetContext(ctx, &author, query, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return author, store.ErrorNotFound
	}
	return author, err
}

// Update modifies an existing author in the database.
func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {
	sets, args := r.prepareArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE authors SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// Delete removes an author by ID from the database.
func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM authors WHERE id=$1 RETURNING id`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// prepareArgs prepares the update arguments for the SQL query.
func (r *AuthorRepository) prepareArgs(data author.Entity) ([]string, []interface{}) {
	var sets []string
	var args []interface{}

	if data.FullName != nil {
		args = append(args, data.FullName)
		sets = append(sets, fmt.Sprintf("full_name=$%d", len(args)))
	}
	if data.Pseudonym != nil {
		args = append(args, data.Pseudonym)
		sets = append(sets, fmt.Sprintf("pseudonym=$%d", len(args)))
	}
	if data.Specialty != nil {
		args = append(args, data.Specialty)
		sets = append(sets, fmt.Sprintf("specialty=$%d", len(args)))
	}

	return sets, args
}
