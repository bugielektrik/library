package fixtures

import (
	"time"

	reservationdomain "library-service/internal/reservations/domain"
)

// ValidReservation returns a valid fulfilled reservation entity for testing
func ValidReservation() reservationdomain.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	fulfilledAt := now.Add(1 * time.Hour)
	return reservationdomain.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013267",
		BookID:      "550e8400-e29b-41d4-a716-446655440000",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservationdomain.StatusFulfilled,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: &fulfilledAt,
		CancelledAt: nil,
	}
}

// PendingReservation returns a reservation with pending status
func PendingReservation() reservationdomain.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return reservationdomain.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013268",
		BookID:      "550e8400-e29b-41d4-a716-446655440001",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservationdomain.StatusPending,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// CancelledReservation returns a reservation with cancelled status
func CancelledReservation() reservationdomain.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	cancelledAt := now.Add(2 * time.Hour)
	return reservationdomain.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013269",
		BookID:      "550e8400-e29b-41d4-a716-446655440002",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservationdomain.StatusCancelled,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: &cancelledAt,
	}
}

// ExpiredReservation returns a reservation that has expired
func ExpiredReservation() reservationdomain.Reservation {
	now := time.Now()
	expiresAt := now.Add(-24 * time.Hour) // Expired yesterday
	return reservationdomain.Reservation{
		ID:          "r4101570-0a35-4dd3-b8f7-745d56013270",
		BookID:      "550e8400-e29b-41d4-a716-446655440003",
		MemberID:    "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:      reservationdomain.StatusExpired,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// ReservationResponse returns a valid reservation response
func ReservationResponse() reservationdomain.Response {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return reservationdomain.Response{
		ID:        "r4101570-0a35-4dd3-b8f7-745d56013267",
		BookID:    "550e8400-e29b-41d4-a716-446655440000",
		MemberID:  "b4101570-0a35-4dd3-b8f7-745d56013263",
		Status:    "active",
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
}

// ReservationResponses returns a slice of reservation responses for testing list operations
func ReservationResponses() []reservationdomain.Response {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	return []reservationdomain.Response{
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
func FulfilledReservation() reservationdomain.Reservation {
	return ValidReservation()
}

// ReservationForCreate returns a reservation entity suitable for repository creation (no ID)
func ReservationForCreate(bookID, memberID string) reservationdomain.Reservation {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	return reservationdomain.Reservation{
		BookID:      bookID,
		MemberID:    memberID,
		Status:      reservationdomain.StatusPending,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		FulfilledAt: nil,
		CancelledAt: nil,
	}
}

// Reservations returns a collection of sample reservations for batch testing
func Reservations() []reservationdomain.Reservation {
	return []reservationdomain.Reservation{
		PendingReservation(),
		FulfilledReservation(),
		ExpiredReservation(),
		CancelledReservation(),
	}
}
