package reservation

import (
	"time"
)

// Status represents the status of a reservation in the system.
type Status string

const (
	// StatusPending means the reservation is waiting for the book to become available
	StatusPending Status = "pending"
	// StatusFulfilled means the book is available and member has been notified
	StatusFulfilled Status = "fulfilled"
	// StatusCancelled means the reservation was cancelled by the member
	StatusCancelled Status = "cancelled"
	// StatusExpired means the reservation expired without being fulfilled
	StatusExpired Status = "expired"
)

// Reservation represents a book reservation in the system.
type Reservation struct {
	// ID is the unique identifier for the reservation.
	ID string `db:"id" bson:"_id"`

	// BookID is the ID of the book being reserved.
	BookID string `db:"book_id" bson:"book_id"`

	// MemberID is the ID of the member making the reservation.
	MemberID string `db:"member_id" bson:"member_id"`

	// Status is the current status of the reservation.
	Status Status `db:"status" bson:"status"`

	// CreatedAt is the timestamp when the reservation was created.
	CreatedAt time.Time `db:"created_at" bson:"created_at"`

	// ExpiresAt is the timestamp when the reservation will expire if not fulfilled.
	ExpiresAt time.Time `db:"expires_at" bson:"expires_at"`

	// FulfilledAt is the timestamp when the reservation was fulfilled (book became available).
	FulfilledAt *time.Time `db:"fulfilled_at" bson:"fulfilled_at"`

	// CancelledAt is the timestamp when the reservation was cancelled.
	CancelledAt *time.Time `db:"cancelled_at" bson:"cancelled_at"`
}

// New creates a new Reservation instance.
func New(req Request) Reservation {
	now := time.Now()
	// Default expiration is 7 days from creation
	expiresAt := now.Add(7 * 24 * time.Hour)

	return Reservation{
		BookID:    req.BookID,
		MemberID:  req.MemberID,
		Status:    StatusPending,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
}

// IsActive returns true if the reservation is in an active state (pending or fulfilled).
func (r Reservation) IsActive() bool {
	return r.Status == StatusPending || r.Status == StatusFulfilled
}

// IsPending returns true if the reservation is still pending.
func (r Reservation) IsPending() bool {
	return r.Status == StatusPending
}

// IsExpired returns true if the reservation has expired based on current time.
func (r Reservation) IsExpired() bool {
	return r.Status == StatusPending && time.Now().After(r.ExpiresAt)
}

// CanBeCancelled returns true if the reservation can be cancelled.
func (r Reservation) CanBeCancelled() bool {
	return r.Status == StatusPending || r.Status == StatusFulfilled
}
