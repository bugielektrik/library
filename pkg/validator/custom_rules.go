package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidations registers custom validation rules
func (v *Validator) RegisterCustomValidations() {
	v.validate.RegisterValidation("isbn", validateISBN)
	v.validate.RegisterValidation("phone", validatePhone)
}

// validateISBN validates ISBN format
func validateISBN(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	// Simple ISBN-10 or ISBN-13 validation
	isbnRegex := regexp.MustCompile(`^(?:ISBN(?:-1[03])?:? )?(?=[0-9X]{10}$|(?=(?:[0-9]+[- ]){3})[- 0-9X]{13}$|97[89][0-9]{10}$|(?=(?:[0-9]+[- ]){4})[- 0-9]{17}$)(?:97[89][- ]?)?[0-9]{1,5}[- ]?[0-9]+[- ]?[0-9]+[- ]?[0-9X]$`)
	return isbnRegex.MatchString(isbn)
}

// validatePhone validates phone number format
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Simple phone validation (international format)
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}
