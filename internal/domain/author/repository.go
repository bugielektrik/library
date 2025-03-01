package author

import (
	"context"
)

// Repository defines the interface for author repository operations.
type Repository interface {
	// List retrieves all author entities.
	List(ctx context.Context) ([]Entity, error)

	// Add inserts a new author entity and returns its ID.
	Add(ctx context.Context, data Entity) (string, error)

	// Get retrieves an author entity by its ID.
	Get(ctx context.Context, id string) (Entity, error)

	// Update modifies an existing author entity by its ID.
	Update(ctx context.Context, id string, data Entity) error

	// Delete removes an author entity by its ID.
	Delete(ctx context.Context, id string) error
}
