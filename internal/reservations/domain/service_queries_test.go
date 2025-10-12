package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_GetNextPendingReservation(t *testing.T) {
	service := NewService()

	now := time.Now()
	earlier := now.Add(-2 * time.Hour)
	latest := now.Add(-1 * time.Hour)

	tests := []struct {
		name         string
		reservations []Reservation
		wantID       string
		wantNil      bool
	}{
		{
			name: "returns oldest pending reservation",
			reservations: []Reservation{
				{ID: "res-3", Status: StatusPending, CreatedAt: latest},
				{ID: "res-1", Status: StatusPending, CreatedAt: earlier},
				{ID: "res-2", Status: StatusPending, CreatedAt: now},
			},
			wantID: "res-1",
		},
		{
			name: "ignores non-pending reservations",
			reservations: []Reservation{
				{ID: "res-1", Status: StatusFulfilled, CreatedAt: earlier},
				{ID: "res-2", Status: StatusPending, CreatedAt: now},
				{ID: "res-3", Status: StatusCancelled, CreatedAt: latest},
			},
			wantID: "res-2",
		},
		{
			name: "returns nil when no pending reservations",
			reservations: []Reservation{
				{ID: "res-1", Status: StatusFulfilled, CreatedAt: earlier},
				{ID: "res-2", Status: StatusCancelled, CreatedAt: now},
			},
			wantNil: true,
		},
		{
			name:         "returns nil for empty list",
			reservations: []Reservation{},
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetNextPendingReservation(tt.reservations)

			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantID, result.ID)
			}
		})
	}
}

func TestService_CalculateExpirationDate(t *testing.T) {
	service := NewService()

	now := time.Now()

	tests := []struct {
		name            string
		createdAt       time.Time
		daysUntilExpiry int
		wantDays        int
	}{
		{
			name:            "custom expiration period",
			createdAt:       now,
			daysUntilExpiry: 14,
			wantDays:        14,
		},
		{
			name:            "default 7 days for zero value",
			createdAt:       now,
			daysUntilExpiry: 0,
			wantDays:        7,
		},
		{
			name:            "default 7 days for negative value",
			createdAt:       now,
			daysUntilExpiry: -5,
			wantDays:        7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateExpirationDate(tt.createdAt, tt.daysUntilExpiry)
			expected := tt.createdAt.Add(time.Duration(tt.wantDays) * 24 * time.Hour)

			assert.Equal(t, expected, result)
		})
	}
}
