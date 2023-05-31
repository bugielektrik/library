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

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Repository) error

// Repository is an implementation of the Repository
type Repository struct {
	postgres *sqlx.DB

	Author author.Repository
	Book   book.Repository
	Member member.Repository
}

// New takes a variable amount of Configuration functions and returns a new Repository
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (s *Repository, err error) {
	// Create the repository
	s = &Repository{}

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
	if r.postgres != nil {
		r.postgres.Close()
	}
}

// WithMemoryDatabase applies a memory database to the Repository
func WithMemoryDatabase() Configuration {
	return func(s *Repository) (err error) {
		// Create the memory database, if we needed parameters, such as connection strings they could be inputted here
		s.Author = memory.NewAuthorRepository()
		s.Book = memory.NewBookRepository()
		s.Member = memory.NewMemberRepository()

		return
	}
}

// WithPostgresDatabase applies a postgres database to the Repository
func WithPostgresDatabase(dataSourceName string) Configuration {
	return func(s *Repository) (err error) {
		// Create the postgres database, if we needed parameters, such as connection strings they could be inputted here
		s.postgres, err = database.New(dataSourceName)
		if err != nil {
			return
		}

		err = database.Migrate(dataSourceName)
		if err != nil {
			return
		}

		s.Author = postgres.NewAuthorRepository(s.postgres)
		s.Book = postgres.NewBookRepository(s.postgres)
		s.Member = postgres.NewMemberRepository(s.postgres)

		return
	}
}
