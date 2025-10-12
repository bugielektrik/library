package errors

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// FromError converts a domain error to ErrorResponse
func FromError(err error) ErrorResponse {
	if domainErr, ok := err.(*Error); ok {
		return ErrorResponse{
			Error: ErrorDetail{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Details: domainErr.Details,
			},
		}
	}

	// Fallback for non-domain errors
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		},
	}
}

// ValidationError represents validation error details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error ValidationErrorDetail `json:"error"`
}

// ValidationErrorDetail contains validation error information
type ValidationErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(validationErrors []ValidationError) ValidationErrorResponse {
	return ValidationErrorResponse{
		Error: ValidationErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Errors:  validationErrors,
		},
	}
}
