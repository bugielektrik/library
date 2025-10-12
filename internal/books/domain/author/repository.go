package author

import (
	"context"
)

// Repository defines the interface for author repository service.
type Repository interface {
	// List retrieves all authors.
	List(ctx context.Context) ([]Author, error)

	// Add inserts a new author and returns its ID.
	Add(ctx context.Context, data Author) (string, error)

	// Get retrieves an author by its ID.
	Get(ctx context.Context, id string) (Author, error)

	// Update modifies an existing author by its ID.
	Update(ctx context.Context, id string, data Author) error

	// Delete removes an author by its ID.
	Delete(ctx context.Context, id string) error
}
