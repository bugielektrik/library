package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"library-service/internal/domain/author"
	"library-service/internal/repository/sqlc"
	"library-service/pkg/store"
)

type AuthorRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
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
			FullName:  &dbAuthor.FullName,
			Pseudonym: &dbAuthor.Pseudonym,
			Specialty: &dbAuthor.Specialty,
		})
	}

	return authors, nil
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	author, err := r.queries.AddAuthor(ctx, sqlc.AddAuthorParams{
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
	})
	if err != nil {
		return "", nil
	}
	return author, nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {
	dbAuthor, err := r.queries.GetAuthor(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return author.Entity{}, store.ErrorNotFound
		}
		return author.Entity{}, err
	}
	authorEntity := author.Entity{
		ID:        dbAuthor.ID,
		FullName:  &dbAuthor.FullName,
		Pseudonym: &dbAuthor.Pseudonym,
		Specialty: &dbAuthor.Specialty,
	}
	return authorEntity, nil
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {
	_, err := r.queries.UpdateAuthor(ctx, sqlc.UpdateAuthorParams{
		ID:        id,
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.DeleteAuthor(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrorNotFound
		}
		return err
	}
	return nil
}
