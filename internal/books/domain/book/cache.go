package book

import "context"

// Cache defines the interface for book cache service.
type Cache interface {
	// Get retrieves a book by its ID from the cache.
	Get(ctx context.Context, id string) (Book, error)

	// Set stores a book in the cache.
	Set(ctx context.Context, id string, book Book) error
}
