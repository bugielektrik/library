package member

import (
	"context"
)

// Repository defines the interface for member repository operations.
type Repository interface {
	// List retrieves all member entities.
	List(ctx context.Context) ([]Entity, error)

	// Add inserts a new member entity and returns its ID.
	Add(ctx context.Context, data Entity) (string, error)

	// Get retrieves a member entity by its ID.
	Get(ctx context.Context, id string) (Entity, error)

	// Update modifies an existing member entity identified by its ID.
	Update(ctx context.Context, id string, data Entity) error

	// Delete removes a member entity by its ID.
	Delete(ctx context.Context, id string) error
}
