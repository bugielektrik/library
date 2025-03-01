package book

import "context"

// Repository defines the interface for book repository operations.
type Repository interface {
	// List retrieves all book entities.
	List(ctx context.Context) ([]Entity, error)

	// Add inserts a new book entity and returns its ID.
	Add(ctx context.Context, data Entity) (string, error)

	// Get retrieves a book entity by its ID.
	Get(ctx context.Context, id string) (Entity, error)

	// Update modifies an existing book entity by its ID.
	Update(ctx context.Context, id string, data Entity) error

	// Delete removes a book entity by its ID.
	Delete(ctx context.Context, id string) error
}
