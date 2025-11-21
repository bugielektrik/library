package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/author"
	"library-service/internal/repository/sqlc"
	"library-service/pkg/store"
)

type AuthorRepository struct {
	db      *sqlx.DB
	queries *sqlc.Queries
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Entity, error) {
	dbAuthors, err := r.queries.ListAuthors(ctx)
	if err != nil {
		return nil, err
	}

	authors := make([]author.Entity, 0, len(dbAuthors))
	for _, dbAuthor := range dbAuthors {
		authors = append(authors, author.Entity{
			ID:        dbAuthor.ID,
			FullName:  nullStringToPtr(dbAuthor.FullName),
			Pseudonym: nullStringToPtr(dbAuthor.Pseudonym),
			Specialty: nullStringToPtr(dbAuthor.Specialty),
		})
	}

	return authors, nil
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	author, err := r.queries.AddAuthor(ctx, sqlc.AddAuthorParams{
		FullName:  ptrToNullString(data.FullName),
		Pseudonym: ptrToNullString(data.Pseudonym),
		Specialty: ptrToNullString(data.Specialty),
	})
	if err != nil {
		return "", nil
	}
	return author, nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {
	dbAuthor, err := r.queries.GetAuthor(ctx, id)
	if err != nil {
		return author.Entity{}, err
	}
	author := author.Entity{
		ID:        dbAuthor.ID,
		FullName:  &dbAuthor.FullName.String,
		Pseudonym: &dbAuthor.Pseudonym.String,
		Specialty: &dbAuthor.Specialty.String,
	}
	return author, nil
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {
	_, err := r.queries.UpdateAuthor(ctx, sqlc.UpdateAuthorParams{
		ID:        id,
		FullName:  ptrToNullString(data.FullName),
		Pseudonym: ptrToNullString(data.Pseudonym),
		Specialty: ptrToNullString(data.Specialty),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
	_, err := r.queries.DeleteAuthor(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func ptrToNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
