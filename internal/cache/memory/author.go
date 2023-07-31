package memory

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"

	"library-service/internal/domain/author"
)

type AuthorCache struct {
	cache      *cache.Cache
	repository author.Repository
}

func NewAuthorCache(r author.Repository) *AuthorCache {
	c := cache.New(5*time.Minute, 10*time.Minute) // Cache with 5 minutes expiration and 10 minutes cleanup interval
	return &AuthorCache{
		cache:      c,
		repository: r,
	}
}

func (c *AuthorCache) Get(ctx context.Context, id string) (dest author.Entity, err error) {
	// Check if data is available in the cache
	if data, found := c.cache.Get(id); found {
		// Data found in the cache, return it
		return data.(author.Entity), nil
	}

	// Data not found in the cache, retrieve it from the data source
	dest, err = c.repository.Get(ctx, id)
	if err != nil {
		return
	}

	// Store the retrieved data in the cache for future use
	c.cache.Set(id, dest, cache.DefaultExpiration)

	return
}
