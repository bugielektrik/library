package domain

import (
	"errors"
	"net/http"
	"time"
)

// Request represents the request payload for creating a reservation.
type Request struct {
	BookID   string `json:"book_id"`
	MemberID string `json:"member_id"`
}

// Bind validates the request payload.
func (r *Request) Bind(req *http.Request) error {
	if r.BookID == "" {
		return errors.New("book_id: cannot be blank")
	}

	if r.MemberID == "" {
		return errors.New("member_id: cannot be blank")
	}

	return nil
}

// Response represents the response payload for reservation service.
type Response struct {
	ID          string     `json:"id"`
	BookID      string     `json:"book_id"`
	MemberID    string     `json:"member_id"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	FulfilledAt *time.Time `json:"fulfilled_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
}

// ParseFromReservation converts a reservation entity to a response payload.
func ParseFromReservation(data Reservation) Response {
	return Response{
		ID:          data.ID,
		BookID:      data.BookID,
		MemberID:    data.MemberID,
		Status:      data.Status,
		CreatedAt:   data.CreatedAt,
		ExpiresAt:   data.ExpiresAt,
		FulfilledAt: data.FulfilledAt,
		CancelledAt: data.CancelledAt,
	}
}

// ParseFromReservations converts a list of reservations to a list of response payloads.
func ParseFromReservations(data []Reservation) []Response {
	res := make([]Response, len(data))
	for i, reservation := range data {
		res[i] = ParseFromReservation(reservation)
	}
	return res
}
