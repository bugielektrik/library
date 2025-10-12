package domain

import (
	"library-service/internal/pkg/errors"
	"time"
)

// Service encapsulates business logic for reservations that doesn't naturally
// belong to a single entity. This is a domain service in DDD terms.
//
// Key Responsibilities:
//   - Reservation eligibility (can member reserve book?)
//   - Status transitions (pending → fulfilled/cancelled/expired)
//   - Cross-entity validation (checking member's borrowed books)
//   - Expiration date calculation
//
// See Also:
//   - Use case example: internal/usecase/reservationops/create_reservation.go (demonstrates cross-domain validation)
//   - Similar service: internal/books/domain/book/service.go (comprehensive example), internal/domain/payment/service.go
//   - ADR: .claude/adr/003-domain-service-vs-infrastructure.md (pure business logic pattern)
//   - ADR: .claude/adr/002-clean-architecture-boundaries.md (domain layer rules)
//   - Test: internal/usecase/reservationops/create_reservation_test.go
type Service struct {
	// Domain service are typically stateless
	// If state is needed, it should be passed as parameters
}

// NewService creates a new reservation domain service
func NewService() *Service {
	return &Service{}
}

// ValidateReservation validates reservation entity according to business rules
func (s *Service) Validate(reservation Reservation) error {
	if reservation.BookID == "" {
		return errors.ErrValidation.WithDetails("field", "book_id").WithDetails("reason", "book_id is required")
	}

	if reservation.MemberID == "" {
		return errors.ErrValidation.WithDetails("field", "member_id").WithDetails("reason", "member_id is required")
	}

	if reservation.ExpiresAt.Before(reservation.CreatedAt) {
		return errors.ErrValidation.WithDetails("field", "expires_at").WithDetails("reason", "expiration date must be after creation date")
	}

	return nil
}

// CanMemberReserveBook checks if a member can reserve a book
// Business rules:
// - A member cannot reserve the same book multiple times with active reservations
// - A member cannot reserve a book they currently have borrowed
func (s *Service) CanMemberReserveBook(memberID, bookID string, existingReservations []Reservation, memberBorrowedBooks []string) error {
	if memberID == "" {
		return errors.ErrValidation.WithDetails("reason", "member_id is required")
	}

	if bookID == "" {
		return errors.ErrValidation.WithDetails("reason", "book_id is required")
	}

	// Check if member already has this book borrowed
	for _, borrowedBookID := range memberBorrowedBooks {
		if borrowedBookID == bookID {
			return errors.ErrValidation.WithDetails("reason", "member already has this book borrowed")
		}
	}

	// Check if member already has an active reservation for this book
	for _, res := range existingReservations {
		if res.MemberID == memberID && res.BookID == bookID && res.IsActive() {
			return errors.ErrAlreadyExists.WithDetails("reason", "member already has an active reservation for this book")
		}
	}

	return nil
}

// CanReservationBeCancelled checks if a reservation can be cancelled
// Business rule: Only pending or fulfilled reservations can be cancelled
func (s *Service) CanReservationBeCancelled(reservation Reservation) error {
	if !reservation.CanBeCancelled() {
		return errors.ErrValidation.WithDetails("reason", "only pending or fulfilled reservations can be cancelled").WithDetails("status", string(reservation.Status))
	}

	return nil
}

// MarkAsFulfilled transitions a reservation to fulfilled status
func (s *Service) MarkAsFulfilled(reservation *Reservation) error {
	if reservation.Status != StatusPending {
		return errors.ErrValidation.WithDetails("reason", "only pending reservations can be fulfilled").WithDetails("status", string(reservation.Status))
	}

	now := time.Now()
	reservation.Status = StatusFulfilled
	reservation.FulfilledAt = &now

	return nil
}

// MarkAsCancelled transitions a reservation to cancelled status
func (s *Service) MarkAsCancelled(reservation *Reservation) error {
	if err := s.CanReservationBeCancelled(*reservation); err != nil {
		return err
	}

	now := time.Now()
	reservation.Status = StatusCancelled
	reservation.CancelledAt = &now

	return nil
}

// MarkAsExpired transitions a reservation to expired status
func (s *Service) MarkAsExpired(reservation *Reservation) error {
	if reservation.Status != StatusPending {
		return errors.ErrValidation.WithDetails("reason", "only pending reservations can be expired").WithDetails("status", string(reservation.Status))
	}

	if !reservation.IsExpired() {
		return errors.ErrValidation.WithDetails("reason", "reservation has not yet expired")
	}

	reservation.Status = StatusExpired

	return nil
}

// GetNextPendingReservation returns the oldest pending reservation from a list
// This is used to determine which member should be notified when a book becomes available
func (s *Service) GetNextPendingReservation(reservations []Reservation) *Reservation {
	var oldest *Reservation

	for i := range reservations {
		if !reservations[i].IsPending() {
			continue
		}

		if oldest == nil || reservations[i].CreatedAt.Before(oldest.CreatedAt) {
			oldest = &reservations[i]
		}
	}

	return oldest
}

// CalculateExpirationDate calculates when a reservation should expire
// Default is 7 days from the creation date
func (s *Service) CalculateExpirationDate(createdAt time.Time, daysUntilExpiry int) time.Time {
	if daysUntilExpiry <= 0 {
		daysUntilExpiry = 7 // Default to 7 days
	}

	return createdAt.Add(time.Duration(daysUntilExpiry) * 24 * time.Hour)
}
