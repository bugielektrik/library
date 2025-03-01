package author

import "context"

// Cache defines the interface for author cache operations.
type Cache interface {
	// Get retrieves an author entity by its ID from the cache.
	// Returns the entity and an error if the operation fails.
	Get(ctx context.Context, id string) (Entity, error)

	// Set stores an author entity in the cache.
	// Returns an error if the operation fails.
	Set(ctx context.Context, id string, entity Entity) error
}
