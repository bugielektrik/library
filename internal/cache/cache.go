package cache

import (
	"github.com/redis/go-redis/v9"

	"library/internal/cache/memory"
	"library/internal/domain/author"
	"library/internal/domain/book"
)

type Dependencies struct {
	AuthorRepository author.Repository
	BookRepository   book.Repository
}

// Configuration is an alias for a function that will take in a pointer to a Cache and modify it
type Configuration func(r *Cache) error

// Cache is an implementation of the Cache
type Cache struct {
	redis        *redis.Client
	dependencies Dependencies

	Author author.Cache
	Book   book.Cache
}

// New takes a variable amount of Configuration functions and returns a new Cache
// Each Configuration will be called in the order they are passed in
func New(d Dependencies, configs ...Configuration) (s *Cache, err error) {
	// Create the cache
	s = &Cache{
		dependencies: d,
	}

	// Apply all Configurations passed in
	for _, cfg := range configs {
		// Pass the cache into the configuration function
		if err = cfg(s); err != nil {
			return
		}
	}

	return
}

// Close closes the cache and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server to finish.
func (r *Cache) Close() {
	if r.redis != nil {
		r.redis.Close()
	}
}

// WithMemoryDatabase applies a memory database to the Cache
func WithMemoryDatabase() Configuration {
	return func(s *Cache) (err error) {
		// Create the memory database, if we needed parameters, such as connection strings they could be inputted here
		s.Author = memory.NewAuthorCache(s.dependencies.AuthorRepository)
		s.Book = memory.NewBookCache(s.dependencies.BookRepository)

		return
	}
}

//// WithRedisDatabase applies a postgres database to the Cache
//func WithRedisDatabase(a author.Repository, b book.Repository) Configuration {
//	return func(s *Cache) (err error) {
//		// Create the postgres database, if we needed parameters, such as connection strings they could be inputted here
//		s.redis, err = database.New(dataSourceName)
//		if err != nil {
//			return
//		}
//
//		err = database.Migrate(dataSourceName)
//		if err != nil {
//			return
//		}
//
//		s.Author = redis.NewAuthorCache(a)
//		s.Book = redis.NewBookCache(b)
//
//		return
//	}
//}
