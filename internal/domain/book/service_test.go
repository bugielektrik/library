package book

import (
	"testing"

	"library-service/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestService_ValidateISBN(t *testing.T) {
	service := NewService()

	tests := []struct {
		name      string
		isbn      string
		wantError bool
		errorType *errors.Error
	}{
		{
			name:      "valid ISBN-13",
			isbn:      "978-0134190440",
			wantError: false,
		},
		{
			name:      "valid ISBN-13 without hyphens",
			isbn:      "9780134190440",
			wantError: false,
		},
		{
			name:      "valid ISBN-10",
			isbn:      "0134190440",
			wantError: false,
		},
		{
			name:      "valid ISBN-10 with X checksum",
			isbn:      "043942089X",
			wantError: false,
		},
		{
			name:      "empty ISBN",
			isbn:      "",
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name:      "too short",
			isbn:      "123",
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name:      "invalid characters",
			isbn:      "ABC-DEF-GHIJ",
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name:      "invalid ISBN-13 prefix",
			isbn:      "1234567890123",
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name:      "ISBN-13 with invalid checksum",
			isbn:      "9780134190441", // Last digit should be 0
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateISBN(tt.isbn)

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

func TestService_ValidateBook(t *testing.T) {
	service := NewService()

	validName := "The Go Programming Language"
	validGenre := "Technology"
	validISBN := "978-0134190440"
	emptyString := ""

	tests := []struct {
		name      string
		book      Book
		wantError bool
		errorType *errors.Error
	}{
		{
			name: "valid book",
			book: Book{
				Name:    &validName,
				Genre:   &validGenre,
				ISBN:    &validISBN,
				Authors: []string{"author-1"},
			},
			wantError: false,
		},
		{
			name: "missing name",
			book: Book{
				Name:    nil,
				Genre:   &validGenre,
				ISBN:    &validISBN,
				Authors: []string{"author-1"},
			},
			wantError: true,
			errorType: errors.ErrInvalidBookData,
		},
		{
			name: "empty name",
			book: Book{
				Name:    &emptyString,
				Genre:   &validGenre,
				ISBN:    &validISBN,
				Authors: []string{"author-1"},
			},
			wantError: true,
			errorType: errors.ErrInvalidBookData,
		},
		{
			name: "missing genre",
			book: Book{
				Name:    &validName,
				Genre:   nil,
				ISBN:    &validISBN,
				Authors: []string{"author-1"},
			},
			wantError: true,
			errorType: errors.ErrInvalidBookData,
		},
		{
			name: "missing ISBN",
			book: Book{
				Name:    &validName,
				Genre:   &validGenre,
				ISBN:    nil,
				Authors: []string{"author-1"},
			},
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name: "invalid ISBN format",
			book: Book{
				Name:    &validName,
				Genre:   &validGenre,
				ISBN:    stringPtr("invalid"),
				Authors: []string{"author-1"},
			},
			wantError: true,
			errorType: errors.ErrInvalidISBN,
		},
		{
			name: "no authors",
			book: Book{
				Name:    &validName,
				Genre:   &validGenre,
				ISBN:    &validISBN,
				Authors: []string{},
			},
			wantError: true,
			errorType: errors.ErrInvalidBookData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateBook(tt.book)

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

func TestService_NormalizeISBN(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		isbn     string
		expected string
		wantErr  bool
	}{
		{
			name:     "ISBN-13 with hyphens",
			isbn:     "978-0-134-19044-0",
			expected: "9780134190440",
			wantErr:  false,
		},
		{
			name:     "ISBN-13 without hyphens",
			isbn:     "9780134190440",
			expected: "9780134190440",
			wantErr:  false,
		},
		{
			name:     "ISBN-10 converts to ISBN-13",
			isbn:     "0134190440",
			expected: "9780134190440",
			wantErr:  false,
		},
		{
			name:     "invalid ISBN",
			isbn:     "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.NormalizeISBN(tt.isbn)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestService_CanBookBeDeleted(t *testing.T) {
	service := NewService()

	tests := []struct {
		name      string
		book      Book
		wantError bool
	}{
		{
			name: "book with ID can be checked",
			book: Book{
				ID: "book-123",
			},
			wantError: false,
		},
		{
			name: "book without ID fails",
			book: Book{
				ID: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CanBookBeDeleted(tt.book)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
