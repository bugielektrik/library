package reservation

import (
	"time"
)

// Status represents the status of a reservation in the system.
//
// STATE MACHINE:
//
//	pending → fulfilled (book becomes available)
//	pending → cancelled (member cancels)
//	pending → expired (ExpiresAt reached without fulfillment)
//	fulfilled → cancelled (member no longer wants book)
//
// TERMINAL STATES:
//   - cancelled: Cannot transition to any other state
//   - expired: Cannot transition to any other state
//
// BUSINESS RULES:
//   - Default expiration: 7 days from creation
//   - Member notified when status → fulfilled
//   - Expired reservations cleaned up by background worker
type Status string

const (
	// StatusPending means the reservation is waiting for the book to become available.
	// Initial state for all new reservations.
	// Can transition to: fulfilled, cancelled, expired.
	StatusPending Status = "pending"

	// StatusFulfilled means the book is available and member has been notified.
	// Member should pick up the book within ExpiresAt window.
	// Can transition to: cancelled.
	StatusFulfilled Status = "fulfilled"

	// StatusCancelled means the reservation was cancelled by the member.
	// Terminal state - no further transitions allowed.
	StatusCancelled Status = "cancelled"

	// StatusExpired means the reservation expired without being fulfilled.
	// Occurs when time.Now() > ExpiresAt while status is pending.
	// Terminal state - no further transitions allowed.
	StatusExpired Status = "expired"
)

// Reservation represents a book reservation entity in the library system.
//
// TYPE HIERARCHY:
//   - This is a DOMAIN ENTITY (pure business object)
//   - Has NO external dependencies
//   - Used across all layers: Use Case → Adapter → Infrastructure
//
// PURPOSE:
//   - Allows members to reserve books that are currently unavailable
//   - Member notified when book becomes available
//   - Reservation expires if not fulfilled within time window
//
// FIELD DESIGN DECISIONS:
//
// 1. ID, BookID, MemberID, Status (NOT pointers):
//   - Required fields, never null
//   - ID: UUID v4 format
//   - BookID/MemberID: foreign keys to Book and Member entities
//   - Status: defaults to StatusPending
//
// 2. CreatedAt, ExpiresAt (time.Time, NOT pointer):
//   - Required timestamps, never null
//   - ExpiresAt calculated as CreatedAt + 7 days (configurable)
//   - Used by background worker to expire stale reservations
//
// 3. FulfilledAt, CancelledAt (*time.Time, pointers):
//   - Nullable: nil until status transitions
//   - FulfilledAt set when status → fulfilled
//   - CancelledAt set when status → cancelled
//   - Used for analytics and reporting
//
// STATE TRANSITIONS:
//
//	Managed by reservation.Service:
//	- MarkFulfilled(): pending → fulfilled
//	- Cancel(): pending/fulfilled → cancelled
//	- MarkExpired(): pending → expired (if time.Now() > ExpiresAt)
//
// RELATIONSHIPS:
//   - Belongs to Member (member_id foreign key)
//   - Belongs to Book (book_id foreign key)
//   - One-to-Many: Member has many Reservations
//   - One-to-Many: Book has many Reservations
//
// BUSINESS RULES:
//   - Member can have multiple reservations for different books
//   - Member CANNOT have multiple active reservations for same book
//   - Reservation automatically expires after 7 days if not fulfilled
//   - Fulfilled reservations expire if book not picked up within window
//
// NOT CACHED:
//   - Time-sensitive data (expires, frequent status changes)
//   - Always fetched fresh from database
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
