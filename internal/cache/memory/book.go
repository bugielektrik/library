package memory

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	"library-service/internal/domain/book"
)

type BookCache struct {
	cache      *cache.Cache
	repository book.Repository
}

func NewBookCache(r book.Repository) *BookCache {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &BookCache{
		cache:      c,
		repository: r,
	}
}

func (r *BookCache) Get(ctx context.Context, id string) (book.Entity, error) {
	if data, found := r.cache.Get(id); found {
		return data.(book.Entity), nil
	}

	dest, err := r.repository.Get(ctx, id)
	if err != nil {
		return dest, err
	}

	r.cache.Set(id, dest, cache.DefaultExpiration)

	return dest, nil
}

func (r *BookCache) Set(ctx context.Context, id string, data book.Entity) error {
	r.cache.Set(id, data, cache.DefaultExpiration)
	return nil
}
