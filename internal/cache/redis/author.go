package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"library-service/internal/domain/author"
)

type AuthorCache struct {
	cache      *redis.Client
	repository author.Repository
}

func NewAuthorCache(c *redis.Client, r author.Repository) *AuthorCache {
	return &AuthorCache{
		cache:      c,
		repository: r,
	}
}

func (c *AuthorCache) GetByID(ctx context.Context, id string) (dest author.Entity, err error) {
	// Check if data is available in Redis cache
	data, err := c.cache.Get(ctx, id).Result()
	if err == nil {
		// Data found in cache, unmarshal JSON into struct
		if err = json.Unmarshal([]byte(data), &dest); err != nil {
			return
		}
		return
	}

	// Data not found in cache, retrieve it from the data source
	dest, err = c.repository.GetByID(ctx, id)
	if err != nil {
		return
	}

	// Marshal struct data into JSON and database it in Redis cache
	payload, err := json.Marshal(dest)
	if err != nil {
		return
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return
	}

	return
}
