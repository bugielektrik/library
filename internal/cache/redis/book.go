package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"library-service/internal/domain/book"
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

func (c *BookCache) Get(ctx context.Context, id string) (book.Entity, error) {
	data, err := c.cache.Get(ctx, id).Result()
	if err == nil {
		var dest book.Entity
		if err = json.Unmarshal([]byte(data), &dest); err != nil {
			return dest, err
		}
		return dest, nil
	}

	dest, err := c.repository.Get(ctx, id)
	if err != nil {
		return dest, err
	}

	payload, err := json.Marshal(dest)
	if err != nil {
		return dest, err
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return dest, err
	}

	return dest, nil
}

func (c *BookCache) Set(ctx context.Context, id string, data book.Entity) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, id, payload, 5*time.Minute).Err()
}
