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
	c := cache.New(5*time.Minute, 10*time.Minute) // Cache with 5 minutes expiration and 10 minutes cleanup interval
	return &BookCache{
		cache:      c,
		repository: r,
	}
}

func (r *BookCache) Get(ctx context.Context, id string) (dest book.Entity, err error) {
	// Check if data is available in the cache
	if data, found := r.cache.Get(id); found {
		// Data found in the cache, return it
		return data.(book.Entity), nil
	}

	// Data not found in the cache, retrieve it from the data source
	dest, err = r.repository.Get(ctx, id)
	if err != nil {
		return
	}

	// Store the retrieved data in the cache for future use
	r.cache.Set(id, dest, cache.DefaultExpiration)

	return
}
