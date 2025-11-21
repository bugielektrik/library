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
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &AuthorCache{
		cache:      c,
		repository: r,
	}
}

func (c *AuthorCache) Get(ctx context.Context, id string) (author.Entity, error) {
	if data, found := c.cache.Get(id); found {
		return data.(author.Entity), nil
	}

	entity, err := c.repository.Get(ctx, id)
	if err != nil {
		return author.Entity{}, err
	}

	c.cache.Set(id, entity, cache.DefaultExpiration)

	return entity, nil
}

func (c *AuthorCache) Set(ctx context.Context, id string, entity author.Entity) error {
	c.cache.Set(id, entity, cache.DefaultExpiration)
	return nil
}
