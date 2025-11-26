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

func (c *AuthorCache) Get(ctx context.Context, id string) (author.Entity, error) {
	data, err := c.cache.Get(ctx, id).Result()
	if err == nil {
		var entity author.Entity
		if err = json.Unmarshal([]byte(data), &entity); err != nil {
			return author.Entity{}, err
		}
		return entity, nil
	}

	entity, err := c.repository.Get(ctx, id)
	if err != nil {
		return author.Entity{}, err
	}

	payload, err := json.Marshal(entity)
	if err != nil {
		return author.Entity{}, err
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return author.Entity{}, err
	}

	return entity, nil
}

func (c *AuthorCache) Set(ctx context.Context, id string, entity author.Entity) error {
	payload, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	if err = c.cache.Set(ctx, id, payload, 5*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}
