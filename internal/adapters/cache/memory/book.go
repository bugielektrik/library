package memory

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	"library-service/internal/domain/book"
)

// BookCache handles caching operations for books using an in-memory cache.
type BookCache struct {
	cache      *cache.Cache
	repository book.Repository
}

// NewBookCache creates a new BookCache.
func NewBookCache(r book.Repository) *BookCache {
	c := cache.New(5*time.Minute, 10*time.Minute) // Cache with 5 minutes expiration and 10 minutes cleanup interval
	return &BookCache{
		cache:      c,
		repository: r,
	}
}

// Get retrieves a book entity by its ID from the cache.
func (r *BookCache) Get(ctx context.Context, id string) (book.Book, error) {
	// Check if data is available in the cache
	if data, found := r.cache.Get(id); found {
		// Data found in the cache, return it
		return data.(book.Book), nil
	}

	// Data not found in the cache, retrieve it from the data source
	dest, err := r.repository.Get(ctx, id)
	if err != nil {
		return dest, err
	}

	// Store the retrieved data in the cache for future use
	r.cache.Set(id, dest, cache.DefaultExpiration)

	return dest, nil
}

// Set stores a book entity in the cache.
func (r *BookCache) Set(ctx context.Context, id string, data book.Book) error {
	r.cache.Set(id, data, cache.DefaultExpiration)
	return nil
}
