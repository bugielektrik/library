package repository

import (
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/repository/memory"
	"library-service/internal/repository/mongo"
	"library-service/internal/repository/postgres"
	"library-service/pkg/store"
)

// Configuration is an alias for a function that will take in a pointer to a Repository and modify it
type Configuration func(r *Repositories) error

// Repositories is an implementation of the Repositories
type Repositories struct {
	mongo    *store.Mongo
	postgres *store.SQL

	Author author.Repository
	Book   book.Repository
	Member member.Repository
}

// New takes a variable amount of Configuration functions and returns a new Repository
// Each Configuration will be called in the order they are passed in
func New(configs ...Configuration) (s *Repositories, err error) {
	// Create the repository
	s = &Repositories{}

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
func (r *Repositories) Close() {
	if r.postgres.Connection != nil {
		r.postgres.Connection.Close()
	}

	if r.mongo.Connection != nil {
		r.mongo.Connection.Disconnect(nil)
	}
}

// WithMemoryStore applies a memory store to the Repository
func WithMemoryStore() Configuration {
	return func(s *Repositories) (err error) {
		// Create the memory store, if we needed parameters, such as connection strings they could be inputted here
		s.Author = memory.NewAuthorRepository()
		s.Book = memory.NewBookRepository()
		s.Member = memory.NewMemberRepository()

		return
	}
}

// WithMongoStore applies a mongo store to the Repository
func WithMongoStore(uri, name string) Configuration {
	return func(s *Repositories) (err error) {
		// Create the mongo store, if we needed parameters, such as connection strings they could be inputted here
		s.mongo, err = store.NewMongo(uri)
		if err != nil {
			return
		}
		database := s.mongo.Connection.Database(name)

		s.Author = mongo.NewAuthorRepository(database)
		s.Book = mongo.NewBookRepository(database)
		s.Member = mongo.NewMemberRepository(database)

		return
	}
}

// WithPostgresStore applies a postgres store to the Repository
func WithPostgresStore(dataSourceName string) Configuration {
	return func(s *Repositories) (err error) {
		// Create the postgres store, if we needed parameters, such as connection strings they could be inputted here
		s.postgres, err = store.NewSQL(dataSourceName)
		if err != nil {
			return
		}

		if err = store.RunMigrations(dataSourceName); err != nil {
			return
		}

		s.Author = postgres.NewAuthorRepository(s.postgres.Connection)
		s.Book = postgres.NewBookRepository(s.postgres.Connection)
		s.Member = postgres.NewMemberRepository(s.postgres.Connection)

		return
	}
}
