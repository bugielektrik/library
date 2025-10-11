package dto

import (
	"time"

	reservationdomain "library-service/internal/reservations/domain"
)

// CreateReservationRequest represents the request to create a new reservation
type CreateReservationRequest struct {
	BookID string `json:"book_id" validate:"required,uuid4"`
}

// ReservationResponse represents the response for a reservation
type ReservationResponse struct {
	ID          string     `json:"id"`
	BookID      string     `json:"book_id"`
	MemberID    string     `json:"member_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	FulfilledAt *time.Time `json:"fulfilled_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

// FromReservationResponse converts domain reservation.Response to ReservationResponse
func FromReservationResponse(resp reservationdomain.Response) ReservationResponse {
	return ReservationResponse{
		ID:          resp.ID,
		BookID:      resp.BookID,
		MemberID:    resp.MemberID,
		Status:      string(resp.Status),
		CreatedAt:   resp.CreatedAt,
		ExpiresAt:   resp.ExpiresAt,
		FulfilledAt: resp.FulfilledAt,
		CancelledAt: resp.CancelledAt,
	}
}

// FromReservationResponses converts slice of domain reservation.Response to slice of ReservationResponse
func FromReservationResponses(responses []reservationdomain.Response) []ReservationResponse {
	result := make([]ReservationResponse, len(responses))
	for i, resp := range responses {
		result[i] = FromReservationResponse(resp)
	}
	return result
}
