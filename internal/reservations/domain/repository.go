package domain

import "context"

// Repository defines the interface for reservation repository operations.
type Repository interface {
	// Create inserts a new reservation and returns its ID.
	Create(ctx context.Context, reservation Reservation) (string, error)

	// GetByID retrieves a reservation by its ID.
	GetByID(ctx context.Context, id string) (Reservation, error)

	// GetByMemberID retrieves all reservations for a specific member.
	GetByMemberID(ctx context.Context, memberID string) ([]Reservation, error)

	// GetByBookID retrieves all reservations for a specific book.
	GetByBookID(ctx context.Context, bookID string) ([]Reservation, error)

	// GetActiveByMemberAndBook retrieves active reservations for a member and book combination.
	GetActiveByMemberAndBook(ctx context.Context, memberID, bookID string) ([]Reservation, error)

	// Update modifies an existing reservation.
	Update(ctx context.Context, reservation Reservation) error

	// Delete removes a reservation by its ID.
	Delete(ctx context.Context, id string) error

	// ListPending retrieves all pending reservations.
	ListPending(ctx context.Context) ([]Reservation, error)

	// ListExpired retrieves all expired reservations.
	ListExpired(ctx context.Context) ([]Reservation, error)
}
