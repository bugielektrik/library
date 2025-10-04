package cache

import (
	"library-service/internal/adapters/cache/memory"
	"library-service/internal/adapters/cache/redis"
	"library-service/internal/adapters/repository"
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"

	"library-service/internal/infrastructure/store"
)

// Dependencies holds cache dependencies
type Dependencies struct {
	Repositories *repository.Repositories
}

// Configuration function type for cache setup
type Configuration func(*Caches) error

// Caches holds all cache implementations
type Caches struct {
	dependencies Dependencies
	redis        store.Redis

	Author author.Cache
	Book   book.Cache
}

// NewCaches creates a new cache container
func NewCaches(deps Dependencies, configs ...Configuration) (*Caches, error) {
	caches := &Caches{
		dependencies: deps,
	}

	for _, cfg := range configs {
		if err := cfg(caches); err != nil {
			return nil, err
		}
	}

	return caches, nil
}

// Close closes all cache connections
func (c *Caches) Close() {
	if c.redis.Connection != nil {
		c.redis.Connection.Close()
	}
}

// WithMemoryStore configures in-memory caches
func WithMemoryStore() Configuration {
	return func(c *Caches) error {
		c.Author = memory.NewAuthorCache(c.dependencies.Repositories.Author)
		c.Book = memory.NewBookCache(c.dependencies.Repositories.Book)
		return nil
	}
}

// WithRedisStore configures Redis caches
func WithRedisStore(url string) Configuration {
	return func(c *Caches) error {
		rdb, err := store.NewRedis(url)
		if err != nil {
			return err
		}
		c.redis = rdb

		c.Author = redis.NewAuthorCache(rdb.Connection, c.dependencies.Repositories.Author)
		c.Book = redis.NewBookCache(rdb.Connection, c.dependencies.Repositories.Book)

		return nil
	}
}
