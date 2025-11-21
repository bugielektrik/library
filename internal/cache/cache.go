package cache

import (
	"library-service/internal/cache/memory"
	"library-service/internal/cache/redis"
	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/repository"
	"library-service/pkg/store"
)

type Dependencies struct {
	Repositories *repository.Repositories
}

type Configuration func(r *Caches) error

type Caches struct {
	dependencies Dependencies
	redis        store.Redis

	Author author.Cache
	Book   book.Cache
}

func New(dependencies Dependencies, configs ...Configuration) (s *Caches, err error) {
	s = &Caches{
		dependencies: dependencies,
	}

	for _, cfg := range configs {
		if err = cfg(s); err != nil {
			return
		}
	}

	return
}

func (r *Caches) Close() {
	if r.redis.Connection != nil {
		r.redis.Connection.Close()
	}
}

func WithMemoryStore() Configuration {
	return func(s *Caches) (err error) {
		s.Author = memory.NewAuthorCache(s.dependencies.Repositories.Author)
		s.Book = memory.NewBookCache(s.dependencies.Repositories.Book)

		return
	}
}

func WithRedisStore(url string) Configuration {
	return func(s *Caches) (err error) {
		s.redis, err = store.NewRedis(url)
		if err != nil {
			return
		}

		s.Author = redis.NewAuthorCache(s.redis.Connection, s.dependencies.Repositories.Author)
		s.Book = redis.NewBookCache(s.redis.Connection, s.dependencies.Repositories.Book)

		return
	}
}
