package domain

import (
	"library-service/internal/pkg/errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_CanReservationBeCancelled(t *testing.T) {
	service := NewService()

	tests := []struct {
		name        string
		reservation Reservation
		wantError   bool
		errorType   *errors.Error
	}{
		{
			name: "can cancel pending reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusPending,
			},
			wantError: false,
		},
		{
			name: "can cancel fulfilled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusFulfilled,
			},
			wantError: false,
		},
		{
			name: "cannot cancel expired reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusExpired,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "cannot cancel already cancelled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusCancelled,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CanReservationBeCancelled(tt.reservation)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_MarkAsFulfilled(t *testing.T) {
	service := NewService()

	tests := []struct {
		name        string
		reservation Reservation
		wantError   bool
		errorType   *errors.Error
	}{
		{
			name: "can fulfill pending reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusPending,
			},
			wantError: false,
		},
		{
			name: "cannot fulfill already fulfilled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusFulfilled,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "cannot fulfill cancelled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusCancelled,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "cannot fulfill expired reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusExpired,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation := tt.reservation
			err := service.MarkAsFulfilled(&reservation)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StatusFulfilled, reservation.Status)
				assert.NotNil(t, reservation.FulfilledAt)
			}
		})
	}
}

func TestService_MarkAsCancelled(t *testing.T) {
	service := NewService()

	tests := []struct {
		name        string
		reservation Reservation
		wantError   bool
		errorType   *errors.Error
	}{
		{
			name: "can cancel pending reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusPending,
			},
			wantError: false,
		},
		{
			name: "can cancel fulfilled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusFulfilled,
			},
			wantError: false,
		},
		{
			name: "cannot cancel already cancelled reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusCancelled,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "cannot cancel expired reservation",
			reservation: Reservation{
				ID:     "res-1",
				Status: StatusExpired,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation := tt.reservation
			err := service.MarkAsCancelled(&reservation)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StatusCancelled, reservation.Status)
				assert.NotNil(t, reservation.CancelledAt)
			}
		})
	}
}

func TestService_MarkAsExpired(t *testing.T) {
	service := NewService()

	now := time.Now()
	past := now.Add(-8 * 24 * time.Hour)
	future := now.Add(8 * 24 * time.Hour)

	tests := []struct {
		name        string
		reservation Reservation
		wantError   bool
		errorType   *errors.Error
	}{
		{
			name: "can expire past-due pending reservation",
			reservation: Reservation{
				ID:        "res-1",
				Status:    StatusPending,
				CreatedAt: past,
				ExpiresAt: past.Add(7 * 24 * time.Hour),
			},
			wantError: false,
		},
		{
			name: "cannot expire not-yet-expired reservation",
			reservation: Reservation{
				ID:        "res-1",
				Status:    StatusPending,
				CreatedAt: now,
				ExpiresAt: future,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "cannot expire fulfilled reservation",
			reservation: Reservation{
				ID:        "res-1",
				Status:    StatusFulfilled,
				CreatedAt: past,
				ExpiresAt: past.Add(7 * 24 * time.Hour),
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reservation := tt.reservation
			err := service.MarkAsExpired(&reservation)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, StatusExpired, reservation.Status)
			}
		})
	}
}
