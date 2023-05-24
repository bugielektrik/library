package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"library/internal/entity"
	"library/internal/repository/memory"
	"library/internal/repository/postgres"
	"library/pkg/database"
)

type AuthorRepository interface {
	SelectRows(ctx context.Context) (dest []entity.Author, err error)
	CreateRow(ctx context.Context, data entity.Author) (id string, err error)
	GetRow(ctx context.Context, id string) (dest entity.Author, err error)
	UpdateRow(ctx context.Context, id string, data entity.Author) (err error)
	DeleteRow(ctx context.Context, id string) (err error)
}

type BookRepository interface {
	SelectRows(ctx context.Context) (dest []entity.Book, err error)
	CreateRow(ctx context.Context, data entity.Book) (id string, err error)
	GetRow(ctx context.Context, id string) (dest entity.Book, err error)
	UpdateRow(ctx context.Context, id string, data entity.Book) (err error)
	DeleteRow(ctx context.Context, id string) (err error)
}

type MemberRepository interface {
	SelectRows(ctx context.Context) (dest []entity.Member, err error)
	CreateRow(ctx context.Context, data entity.Member) (id string, err error)
	GetRow(ctx context.Context, id string) (dest entity.Member, err error)
	UpdateRow(ctx context.Context, id string, data entity.Member) (err error)
	DeleteRow(ctx context.Context, id string) (err error)
}

type Repository struct {
	postgres *sqlx.DB

	Author AuthorRepository
	Book   BookRepository
	Member MemberRepository
}

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Repository) error

// New takes a variable amount of Configuration functions and returns a new Repository
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (r *Repository, err error) {
	// Create the Repository
	r = &Repository{}
	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the service into the configuration function
		if err = cfg(r); err != nil {
			return
		}
	}
	return
}

func (r Repository) Close() {
	if r.postgres != nil {
		r.postgres.Close()
	}
}

func WithMemoryRepository() Configuration {
	return func(r *Repository) (err error) {
		r.Author = memory.NewAuthorRepository()
		r.Book = memory.NewBookRepository()
		r.Member = memory.NewMemberRepository()
		return
	}
}

func WithPostgresRepository(dataSourceName string) Configuration {
	return func(r *Repository) (err error) {
		r.postgres, err = database.New(dataSourceName)
		if err != nil {
			return
		}

		err = database.Migrate(dataSourceName)
		if err != nil {
			return
		}

		r.Author = postgres.NewAuthorRepository(r.postgres)
		r.Book = postgres.NewBookRepository(r.postgres)
		r.Member = postgres.NewMemberRepository(r.postgres)
		return
	}
}
