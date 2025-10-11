package fixtures

import (
	"time"

	"library-service/internal/domain/reservation"
)

// ValidReservation returns a valid fulfilled reservation entity for testing
func ValidReservation() reservation.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	fulfilledAt := now.Add(1 * time.Hour)
	return reservation.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013267",
		BookID:      "550e8400-e29b-41d4-a716-446655440000",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservation.StatusFulfilled,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: &fulfilledAt,
		CancelledAt: nil,
	}
}

// PendingReservation returns a reservation with pending status
func PendingReservation() reservation.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return reservation.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013268",
		BookID:      "550e8400-e29b-41d4-a716-446655440001",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservation.StatusPending,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// CancelledReservation returns a reservation with cancelled status
func CancelledReservation() reservation.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	cancelledAt := now.Add(2 * time.Hour)
	return reservation.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013269",
		BookID:      "550e8400-e29b-41d4-a716-446655440002",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservation.StatusCancelled,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: &cancelledAt,
	}
}

// ExpiredReservation returns a reservation that has expired
func ExpiredReservation() reservation.Reservation {
	now := time.Now()
	expiresAt := now.Add(-24 * time.Hour) // Expired yesterday
	return reservation.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013270",
		BookID:      "550e8400-e29b-41d4-a716-446655440003",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservation.StatusExpired,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// ReservationResponse returns a valid reservation response
func ReservationResponse() reservation.Response {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return reservation.Response{
		ID:        "r4101570-0a35-4dd3-b8f7-745d56013267",
		BookID:    "550e8400-e29b-41d4-a716-446655440000",
		MemberID:  "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:    "active",
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
}

// ReservationResponses returns a slice of reservation responses for testing list operations
func ReservationResponses() []reservation.Response {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return []reservation.Response{
		{
			ID:        "r4101570-0a35-4dd3-b8f7-745d56013267",
			BookID:    "550e8400-e29b-41d4-a716-446655440000",
			MemberID:  "b4101570-0a35-4dd3-b8f7-745d56013263",
			Status:    "active",
			ExpiresAt: expiresAt,
			CreatedAt: now,
		},
		{
			ID:        "r4101570-0a35-4dd3-b8f7-745d56013268",
			BookID:    "550e8400-e29b-41d4-a716-446655440001",
			MemberID:  "b4101570-0a35-4dd3-b8f7-745d56013263",
			Status:    "pending",
			ExpiresAt: expiresAt,
			CreatedAt: now,
		},
	}
}

// FulfilledReservation is an alias for ValidReservation for integration tests
func FulfilledReservation() reservation.Reservation {
	return ValidReservation()
}

// ReservationForCreate returns a reservation entity suitable for repository creation (no ID)
func ReservationForCreate(bookID, memberID string) reservation.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	return reservation.Reservation{
		BookID:      bookID,
		MemberID:    memberID,
		Status:      reservation.StatusPending,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// Reservations returns a collection of sample reservations for batch testing
func Reservations() []reservation.Reservation {
	return []reservation.Reservation{
		PendingReservation(),
		FulfilledReservation(),
		ExpiredReservation(),
		CancelledReservation(),
	}
}
