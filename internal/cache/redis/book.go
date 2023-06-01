package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"library/internal/domain/book"
)

type BookCache struct {
	cache      *redis.Client
	repository book.Repository
}

func NewBookCache(c *redis.Client, r book.Repository) *BookCache {
	return &BookCache{
		cache:      c,
		repository: r,
	}
}

func (c *BookCache) Get(ctx context.Context, id string) (dest book.Entity, err error) {
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
	dest, err = c.repository.Get(ctx, id)
	if err != nil {
		return
	}

	// Marshal struct data into JSON and store it in Redis cache
	payload, err := json.Marshal(dest)
	if err != nil {
		return
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return
	}

	return
}
