package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReservation_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"pending is active", StatusPending, true},
		{"fulfilled is active", StatusFulfilled, true},
		{"cancelled is not active", StatusCancelled, false},
		{"expired is not active", StatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reservation{Status: tt.status}
			assert.Equal(t, tt.want, r.IsActive())
		})
	}
}

func TestReservation_IsPending(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"pending returns true", StatusPending, true},
		{"fulfilled returns false", StatusFulfilled, false},
		{"cancelled returns false", StatusCancelled, false},
		{"expired returns false", StatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reservation{Status: tt.status}
			assert.Equal(t, tt.want, r.IsPending())
		})
	}
}

func TestReservation_IsExpired(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name      string
		status    Status
		expiresAt time.Time
		want      bool
	}{
		{"pending past expiration", StatusPending, past, true},
		{"pending before expiration", StatusPending, future, false},
		{"fulfilled past expiration", StatusFulfilled, past, false},
		{"cancelled past expiration", StatusCancelled, past, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reservation{Status: tt.status, ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, r.IsExpired())
		})
	}
}

func TestReservation_CanBeCancelled(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"pending can be cancelled", StatusPending, true},
		{"fulfilled can be cancelled", StatusFulfilled, true},
		{"cancelled cannot be cancelled", StatusCancelled, false},
		{"expired cannot be cancelled", StatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reservation{Status: tt.status}
			assert.Equal(t, tt.want, r.CanBeCancelled())
		})
	}
}
