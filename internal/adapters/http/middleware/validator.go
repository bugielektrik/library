package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"library-service/internal/adapters/http/dto"
	"library-service/pkg/httputil"
)

// Validator wraps the validator instance
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate validates a struct and returns validation errors if any
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateStruct validates a struct and writes error response if validation fails
func (v *Validator) ValidateStruct(w http.ResponseWriter, data interface{}) bool {
	if err := v.Validate(data); err != nil {
		validationErrors := v.parseValidationErrors(err)

		w.Header().Set(httputil.HeaderContentType, httputil.ContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		response := dto.NewValidationErrorResponse(validationErrors)
		json.NewEncoder(w).Encode(response)

		return false
	}
	return true
}

// parseValidationErrors converts validator errors to DTO validation errors
func (v *Validator) parseValidationErrors(err error) []dto.ValidationError {
	var validationErrors []dto.ValidationError

	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validatorErrs {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   e.Field(),
				Message: v.getErrorMessage(e),
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a human-readable error message for a validation error
func (v *Validator) getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid4":
		return "Invalid UUID format"
	case "isbn":
		return "Invalid ISBN format"
	case "e164":
		return "Invalid phone number format"
	default:
		return "Invalid value"
	}
}
