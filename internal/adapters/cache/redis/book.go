package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"library-service/internal/domain/book"
)

// BookCache handles caching operations for books using Redis.
type BookCache struct {
	cache      *redis.Client
	repository book.Repository
}

// NewBookCache creates a new BookCache.
func NewBookCache(c *redis.Client, r book.Repository) *BookCache {
	return &BookCache{
		cache:      c,
		repository: r,
	}
}

// Get retrieves a book entity by its ID from the cache.
func (c *BookCache) Get(ctx context.Context, id string) (book.Book, error) {
	// Check if data is available in Redis cache
	data, err := c.cache.Get(ctx, id).Result()
	if err == nil {
		// Data found in cache, unmarshal JSON into struct
		var dest book.Book
		if err = json.Unmarshal([]byte(data), &dest); err != nil {
			return dest, err
		}
		return dest, nil
	}

	// Data not found in cache, retrieve it from the data source
	dest, err := c.repository.Get(ctx, id)
	if err != nil {
		return dest, err
	}

	// Marshal struct data into JSON and store it in Redis cache
	payload, err := json.Marshal(dest)
	if err != nil {
		return dest, err
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return dest, err
	}

	return dest, nil
}

// Set stores a book entity in the cache.
func (c *BookCache) Set(ctx context.Context, id string, data book.Book) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, id, payload, 5*time.Minute).Err()
}
