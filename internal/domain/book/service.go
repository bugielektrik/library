package book

import (
	"regexp"
	"strings"

	"library-service/pkg/errors"
)

// Service encapsulates business logic for books that doesn't naturally
// belong to a single entity. This is a domain service in DDD terms.
type Service struct {
	// Domain services are typically stateless
	// If state is needed, it should be passed as parameters
}

// NewService creates a new book domain service
func NewService() *Service {
	return &Service{}
}

// ValidateISBN validates that an ISBN is in the correct format
// Supports both ISBN-10 and ISBN-13 formats
func (s *Service) ValidateISBN(isbn string) error {
	if isbn == "" {
		return errors.ErrInvalidISBN.WithDetails("reason", "ISBN cannot be empty")
	}

	// Remove hyphens and spaces for validation
	cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	// ISBN-10: 10 digits with possible 'X' at end
	isbn10Pattern := regexp.MustCompile(`^\d{9}[\dX]$`)
	// ISBN-13: 13 digits starting with 978 or 979
	isbn13Pattern := regexp.MustCompile(`^(978|979)\d{10}$`)

	if !isbn10Pattern.MatchString(cleanISBN) && !isbn13Pattern.MatchString(cleanISBN) {
		return errors.ErrInvalidISBN.WithDetails("reason", "ISBN must be 10 or 13 digits in valid format")
	}

	// Validate checksum for ISBN-13
	if len(cleanISBN) == 13 {
		if !s.validateISBN13Checksum(cleanISBN) {
			return errors.ErrInvalidISBN.WithDetails("reason", "Invalid ISBN-13 checksum")
		}
	}

	// Validate checksum for ISBN-10
	if len(cleanISBN) == 10 {
		if !s.validateISBN10Checksum(cleanISBN) {
			return errors.ErrInvalidISBN.WithDetails("reason", "Invalid ISBN-10 checksum")
		}
	}

	return nil
}

// ValidateBook validates book entity according to business rules
func (s *Service) ValidateBook(book Book) error {
	if book.Name == nil || *book.Name == "" {
		return errors.ErrInvalidBookData.WithDetails("field", "name")
	}

	if book.Genre == nil || *book.Genre == "" {
		return errors.ErrInvalidBookData.WithDetails("field", "genre")
	}

	if book.ISBN == nil || *book.ISBN == "" {
		return errors.ErrInvalidISBN.WithDetails("reason", "ISBN is required")
	}

	// Validate ISBN format
	if err := s.ValidateISBN(*book.ISBN); err != nil {
		return err
	}

	if len(book.Authors) == 0 {
		return errors.ErrInvalidBookData.WithDetails("field", "authors")
	}

	return nil
}

// CanBookBeDeleted checks if a book can be safely deleted
// Business rule: A book cannot be deleted if it has active loans or reservations
// For now, this is a placeholder - in a real system, this would check against
// a loans/reservations repository
func (s *Service) CanBookBeDeleted(book Book) error {
	// Placeholder implementation
	// In production, this would check:
	// - No active loans for this book
	// - No pending reservations
	// - Not referenced in any historical records that must be preserved

	if book.ID == "" {
		return errors.ErrInvalidBookData.WithDetails("reason", "Book ID is required")
	}

	// Future: Check against loans repository
	// if hasActiveLoans {
	//     return errors.ErrBookHasActiveLoans
	// }

	return nil
}

// NormalizeISBN normalizes an ISBN to a standard format (ISBN-13 without hyphens)
func (s *Service) NormalizeISBN(isbn string) (string, error) {
	if err := s.ValidateISBN(isbn); err != nil {
		return "", err
	}

	// Remove hyphens and spaces
	cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	// Convert ISBN-10 to ISBN-13 format
	if len(cleanISBN) == 10 {
		// ISBN-10 to ISBN-13: prefix with 978 and recalculate checksum
		isbn13 := "978" + cleanISBN[:9]
		checksum := s.calculateISBN13Checksum(isbn13)
		return isbn13 + string(rune('0'+checksum)), nil
	}

	return cleanISBN, nil
}

// validateISBN13Checksum validates the checksum digit of an ISBN-13
func (s *Service) validateISBN13Checksum(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	checksum := s.calculateISBN13Checksum(isbn[:12])
	return isbn[12] == byte('0'+checksum)
}

// calculateISBN13Checksum calculates the checksum digit for an ISBN-13
func (s *Service) calculateISBN13Checksum(isbn string) int {
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(isbn[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checksum := (10 - (sum % 10)) % 10
	return checksum
}

// validateISBN10Checksum validates the checksum digit of an ISBN-10
func (s *Service) validateISBN10Checksum(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(isbn[i] - '0')
		sum += digit * (10 - i)
	}

	// Last character can be 'X' representing 10
	var checkDigit int
	if isbn[9] == 'X' {
		checkDigit = 10
	} else {
		checkDigit = int(isbn[9] - '0')
	}

	sum += checkDigit

	return sum%11 == 0
}
