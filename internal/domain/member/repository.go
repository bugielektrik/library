package member

import (
	"context"
)

// Repository defines the interface for member repository operations.
type Repository interface {
	// List retrieves all members.
	List(ctx context.Context) ([]Member, error)

	// Add inserts a new member and returns its ID.
	Add(ctx context.Context, data Member) (string, error)

	// Get retrieves a member by its ID.
	Get(ctx context.Context, id string) (Member, error)

	// Update modifies an existing member identified by its ID.
	Update(ctx context.Context, id string, data Member) error

	// Delete removes a member by its ID.
	Delete(ctx context.Context, id string) error
}
