package domain

import (
	"context"
	"time"
)

// Repository defines the interface for member repository operations.
type Repository interface {
	// List retrieves all members.
	List(ctx context.Context) ([]Member, error)

	// Add inserts a new member and returns its ID.
	Add(ctx context.Context, data Member) (string, error)

	// Get retrieves a member by its ID.
	Get(ctx context.Context, id string) (Member, error)

	// GetByEmail retrieves a member by their email address.
	GetByEmail(ctx context.Context, email string) (Member, error)

	// Update modifies an existing member identified by its ID.
	Update(ctx context.Context, id string, data Member) error

	// UpdateLastLogin updates the last login timestamp for a member.
	UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error

	// Delete removes a member by its ID.
	Delete(ctx context.Context, id string) error

	// EmailExists checks if an email is already registered.
	EmailExists(ctx context.Context, email string) (bool, error)
}
