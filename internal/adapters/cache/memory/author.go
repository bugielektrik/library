package memory

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	"library-service/internal/domain/author"
)

// AuthorCache handles caching of author entities in memory.
type AuthorCache struct {
	cache      *cache.Cache
	repository author.Repository
}

// NewAuthorCache creates a new AuthorCache.
func NewAuthorCache(r author.Repository) *AuthorCache {
	c := cache.New(5*time.Minute, 10*time.Minute) // Cache with 5 minutes expiration and 10 minutes cleanup interval
	return &AuthorCache{
		cache:      c,
		repository: r,
	}
}

// Get retrieves an author entity by its ID from the cache or repository.
func (c *AuthorCache) Get(ctx context.Context, id string) (author.Entity, error) {
	// Check if data is available in the cache
	if data, found := c.cache.Get(id); found {
		// Data found in the cache, return it
		return data.(author.Entity), nil
	}

	// Data not found in the cache, retrieve it from the repository
	entity, err := c.repository.Get(ctx, id)
	if err != nil {
		return author.Entity{}, err
	}

	// Store the retrieved data in the cache for future use
	c.cache.Set(id, entity, cache.DefaultExpiration)

	return entity, nil
}

// Set stores an author entity in the cache.
func (c *AuthorCache) Set(ctx context.Context, id string, entity author.Entity) error {
	c.cache.Set(id, entity, cache.DefaultExpiration)
	return nil
}
