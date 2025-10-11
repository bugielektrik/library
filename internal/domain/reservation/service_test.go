package reservation

import (
	"testing"
	"time"

	"library-service/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestService_Validate(t *testing.T) {
	service := NewService()

	now := time.Now()
	future := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name        string
		reservation Reservation
		wantError   bool
		errorType   *errors.Error
	}{
		{
			name: "valid reservation",
			reservation: Reservation{
				BookID:    "book-123",
				MemberID:  "member-456",
				Status:    StatusPending,
				CreatedAt: now,
				ExpiresAt: future,
			},
			wantError: false,
		},
		{
			name: "missing book_id",
			reservation: Reservation{
				MemberID:  "member-456",
				Status:    StatusPending,
				CreatedAt: now,
				ExpiresAt: future,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "missing member_id",
			reservation: Reservation{
				BookID:    "book-123",
				Status:    StatusPending,
				CreatedAt: now,
				ExpiresAt: future,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
		{
			name: "expires_at before created_at",
			reservation: Reservation{
				BookID:    "book-123",
				MemberID:  "member-456",
				Status:    StatusPending,
				CreatedAt: future,
				ExpiresAt: now,
			},
			wantError: true,
			errorType: errors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.reservation)

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

func TestService_CanMemberReserveBook(t *testing.T) {
	service := NewService()

	tests := []struct {
		name                 string
		memberID             string
		bookID               string
		existingReservations []Reservation
		memberBorrowedBooks  []string
		wantError            bool
		errorType            *errors.Error
	}{
		{
			name:                 "can reserve book",
			memberID:             "member-1",
			bookID:               "book-1",
			existingReservations: []Reservation{},
			memberBorrowedBooks:  []string{},
			wantError:            false,
		},
		{
			name:                 "missing member_id",
			memberID:             "",
			bookID:               "book-1",
			existingReservations: []Reservation{},
			memberBorrowedBooks:  []string{},
			wantError:            true,
			errorType:            errors.ErrValidation,
		},
		{
			name:                 "missing book_id",
			memberID:             "member-1",
			bookID:               "",
			existingReservations: []Reservation{},
			memberBorrowedBooks:  []string{},
			wantError:            true,
			errorType:            errors.ErrValidation,
		},
		{
			name:                 "member already has book borrowed",
			memberID:             "member-1",
			bookID:               "book-1",
			existingReservations: []Reservation{},
			memberBorrowedBooks:  []string{"book-1", "book-2"},
			wantError:            true,
			errorType:            errors.ErrValidation,
		},
		{
			name:     "member already has active pending reservation",
			memberID: "member-1",
			bookID:   "book-1",
			existingReservations: []Reservation{
				{
					ID:       "res-1",
					MemberID: "member-1",
					BookID:   "book-1",
					Status:   StatusPending,
				},
			},
			memberBorrowedBooks: []string{},
			wantError:           true,
			errorType:           errors.ErrAlreadyExists,
		},
		{
			name:     "member already has active fulfilled reservation",
			memberID: "member-1",
			bookID:   "book-1",
			existingReservations: []Reservation{
				{
					ID:       "res-1",
					MemberID: "member-1",
					BookID:   "book-1",
					Status:   StatusFulfilled,
				},
			},
			memberBorrowedBooks: []string{},
			wantError:           true,
			errorType:           errors.ErrAlreadyExists,
		},
		{
			name:     "can reserve if previous reservation was cancelled",
			memberID: "member-1",
			bookID:   "book-1",
			existingReservations: []Reservation{
				{
					ID:       "res-1",
					MemberID: "member-1",
					BookID:   "book-1",
					Status:   StatusCancelled,
				},
			},
			memberBorrowedBooks: []string{},
			wantError:           false,
		},
		{
			name:     "can reserve if previous reservation expired",
			memberID: "member-1",
			bookID:   "book-1",
			existingReservations: []Reservation{
				{
					ID:       "res-1",
					MemberID: "member-1",
					BookID:   "book-1",
					Status:   StatusExpired,
				},
			},
			memberBorrowedBooks: []string{},
			wantError:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CanMemberReserveBook(tt.memberID, tt.bookID, tt.existingReservations, tt.memberBorrowedBooks)

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
