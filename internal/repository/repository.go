package repository

import (
	"github.com/jmoiron/sqlx"

	"library/internal/domain/author"
	"library/internal/domain/book"
	"library/internal/domain/member"
	"library/internal/repository/memory"
	"library/internal/repository/postgres"
	"library/pkg/database"
)

type Dependencies struct {
	postgres *sqlx.DB

	PostgresDSN string
}

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Repository) error

// Repository is an implementation of the Repository
type Repository struct {
	dependencies Dependencies

	Author author.Repository
	Book   book.Repository
	Member member.Repository
}

// New takes a variable amount of Configuration functions and returns a new Repository
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (s *Repository, err error) {
	// Create the repository
	s = &Repository{
		dependencies: d,
	}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the repository into the configuration function
		if err = cfg(s); err != nil {
			return
		}
	}
	return
}

// Close closes the repository and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server to finish.
func (r *Repository) Close() {
	if r.dependencies.postgres != nil {
		r.dependencies.postgres.Close()
	}
}

// Migrate looks at the currently active migration version
// and will migrate all the way up (applying all up migrations).
func (r *Repository) Migrate() (err error) {
	if r.dependencies.postgres != nil {
		err = database.Migrate(r.dependencies.PostgresDSN)
		if err != nil {
			return
		}
	}

	return
}

// WithMemoryRepository applies a memory repository to the Repository
func WithMemoryRepository() Configuration {
	return func(s *Repository) (err error) {
		// Create the memory repository, if we needed parameters, such as connection strings they could be inputted here
		s.Author = memory.NewAuthorRepository()
		s.Book = memory.NewBookRepository()
		s.Member = memory.NewMemberRepository()
		return
	}
}

// WithPostgresRepository applies a postgres repository to the Repository
func WithPostgresRepository() Configuration {
	return func(s *Repository) (err error) {
		// Create the postgres repository, if we needed parameters, such as connection strings they could be inputted here
		s.dependencies.postgres, err = database.New(s.dependencies.PostgresDSN)
		if err != nil {
			return
		}

		s.Author = postgres.NewAuthorRepository(s.dependencies.postgres)
		s.Book = postgres.NewBookRepository(s.dependencies.postgres)
		s.Member = postgres.NewMemberRepository(s.dependencies.postgres)
		return
	}
}
