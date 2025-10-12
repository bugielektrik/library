package author

import (
	errors2 "library-service/internal/pkg/errors"
	"strings"
)

// Service provides business logic for author domain.
// For the Author domain, business logic is minimal as it's primarily a data entity.
type Service struct {
	// Domain service are typically stateless
}

// NewService creates a new author domain service.
func NewService() *Service {
	return &Service{}
}

// Validate performs domain validation on an Author entity.
// This ensures business rules are enforced before persistence.
func (s *Service) Validate(a Author) error {
	// At least one name field must be provided
	if (a.FullName == nil || strings.TrimSpace(*a.FullName) == "") &&
		(a.Pseudonym == nil || strings.TrimSpace(*a.Pseudonym) == "") {
		return errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "author_name").
			WithDetail("reason", "either full_name or pseudonym must be provided").
			Build()
	}

	// Validate FullName length if provided
	if a.FullName != nil {
		fullName := strings.TrimSpace(*a.FullName)
		if len(fullName) > 200 {
			return errors2.NewError(errors2.CodeValidation).
				WithDetail("field", "full_name").
				WithDetail("reason", "full name cannot exceed 200 characters").
				WithDetail("length", len(fullName)).
				Build()
		}
	}

	// Validate Pseudonym length if provided
	if a.Pseudonym != nil {
		pseudonym := strings.TrimSpace(*a.Pseudonym)
		if len(pseudonym) > 100 {
			return errors2.NewError(errors2.CodeValidation).
				WithDetail("field", "pseudonym").
				WithDetail("reason", "pseudonym cannot exceed 100 characters").
				WithDetail("length", len(pseudonym)).
				Build()
		}
	}

	// Validate Specialty length if provided
	if a.Specialty != nil {
		specialty := strings.TrimSpace(*a.Specialty)
		if len(specialty) > 100 {
			return errors2.NewError(errors2.CodeValidation).
				WithDetail("field", "specialty").
				WithDetail("reason", "specialty cannot exceed 100 characters").
				WithDetail("length", len(specialty)).
				Build()
		}
	}

	return nil
}

// ValidateUpdate validates an author update request.
// This ensures at least one field is being updated.
func (s *Service) ValidateUpdate(a Author) error {
	// First perform standard validation
	if err := s.Validate(a); err != nil {
		return err
	}

	// For updates, at least one field should be non-nil
	// (This is already enforced by Validate requiring at least one name)

	return nil
}

// GetDisplayName returns the preferred display name for an author.
// Priority: Pseudonym > FullName > "Unknown Author"
func (s *Service) GetDisplayName(a Author) string {
	if a.Pseudonym != nil && strings.TrimSpace(*a.Pseudonym) != "" {
		return strings.TrimSpace(*a.Pseudonym)
	}

	if a.FullName != nil && strings.TrimSpace(*a.FullName) != "" {
		return strings.TrimSpace(*a.FullName)
	}

	return "Unknown Author"
}

// GetSearchTerms returns all searchable terms for an author.
// Useful for search and filtering service.
func (s *Service) GetSearchTerms(a Author) []string {
	terms := make([]string, 0, 3)

	if a.FullName != nil && strings.TrimSpace(*a.FullName) != "" {
		terms = append(terms, strings.ToLower(strings.TrimSpace(*a.FullName)))
	}

	if a.Pseudonym != nil && strings.TrimSpace(*a.Pseudonym) != "" {
		terms = append(terms, strings.ToLower(strings.TrimSpace(*a.Pseudonym)))
	}

	if a.Specialty != nil && strings.TrimSpace(*a.Specialty) != "" {
		terms = append(terms, strings.ToLower(strings.TrimSpace(*a.Specialty)))
	}

	return terms
}
