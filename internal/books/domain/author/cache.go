package author

import "context"

// Cache defines the interface for author cache service.
type Cache interface {
	// Get retrieves an author by its ID from the cache.
	// Returns the author and an error if the operation fails.
	Get(ctx context.Context, id string) (Author, error)

	// Set stores an author in the cache.
	// Returns an error if the operation fails.
	Set(ctx context.Context, id string, author Author) error
}
