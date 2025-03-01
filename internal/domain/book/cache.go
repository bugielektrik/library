package book

import "context"

// Cache defines the interface for book cache operations.
type Cache interface {
	// Get retrieves a book entity by its ID from the cache.
	Get(ctx context.Context, id string) (Entity, error)

	// Set stores a book entity in the cache.
	Set(ctx context.Context, id string, entity Entity) error
}
