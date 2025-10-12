package domain

import (
	"library-service/internal/pkg/errors"
	"testing"
	"time"

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
