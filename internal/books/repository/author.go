package repository

import (
	"context"
	"fmt"
	postgres2 "library-service/internal/pkg/repository/postgres"
	"strings"

	"github.com/jmoiron/sqlx"

	"library-service/internal/books/domain/author"
)

// AuthorRepository handles CRUD operations for authors in a PostgreSQL store.
type AuthorRepository struct {
	postgres2.BaseRepository[author.Author]
}

// Compile-time check that AuthorRepository implements author.Repository
var _ author.Repository = (*AuthorRepository)(nil)

// NewAuthorRepository creates a new AuthorRepository.
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		BaseRepository: postgres2.NewBaseRepository[author.Author](db, "authors"),
	}
}

// List is inherited from BaseRepository

// Add inserts a new author into the store.
func (r *AuthorRepository) Add(ctx context.Context, data author.Author) (string, error) {
	query := `INSERT INTO authors (full_name, pseudonym, specialty) VALUES ($1, $2, $3) RETURNING id`
	args := []interface{}{data.FullName, data.Pseudonym, data.Specialty}
	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return id, postgres2.HandleSQLError(err)
}

// Get is inherited from BaseRepository

// Update modifies an existing author in the store.
func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Author) error {
	sets, args := postgres2.PrepareUpdateArgs(data)
	if len(args) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE authors SET %s, updated_at=CURRENT_TIMESTAMP WHERE id=$%d RETURNING id", strings.Join(sets, ", "), len(args))
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return postgres2.HandleSQLError(err)
}

// Delete is inherited from BaseRepository
