package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/books/domain/author"
)

// AuthorRepository handles CRUD operations for authors in a PostgreSQL store.
type AuthorRepository struct {
	postgres.BaseRepository[author.Author]
}

// NewAuthorRepository creates a new AuthorRepository.
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		BaseRepository: postgres.NewBaseRepository[author.Author](db, "authors"),
	}
}

// List is inherited from BaseRepository

// Add inserts a new author into the store.
func (r *AuthorRepository) Add(ctx context.Context, data author.Author) (string, error) {
	query := `INSERT INTO authors (full_name, pseudonym, specialty) VALUES ($1, $2, $3) RETURNING id`
	args := []interface{}{data.FullName, data.Pseudonym, data.Specialty}
	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return id, postgres.HandleSQLError(err)
}

// Get is inherited from BaseRepository

// Update modifies an existing author in the store.
func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Author) error {
	sets, args := r.prepareArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE authors SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return postgres.HandleSQLError(err)
}

// Delete is inherited from BaseRepository

// prepareArgs prepares the update arguments for the SQL query.
func (r *AuthorRepository) prepareArgs(data author.Author) ([]string, []interface{}) {
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
