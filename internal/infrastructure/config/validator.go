package config

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// ConfigValidator validates configuration values
type ConfigValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new configuration validator
func NewValidator() *ConfigValidator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("url", validateURL)
	v.RegisterValidation("email", validateEmail)
	v.RegisterValidation("port", validatePort)
	v.RegisterValidation("duration", validateDuration)
	v.RegisterValidation("path", validatePath)
	v.RegisterValidation("secret", validateSecret)

	// Use JSON tag names for errors
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &ConfigValidator{
		validator: v,
	}
}

// Validate validates the configuration
func (cv *ConfigValidator) Validate(config *Config) error {
	if err := cv.validator.Struct(config); err != nil {
		return cv.formatValidationError(err)
	}

	// Custom validation logic
	if err := cv.validateCustomRules(config); err != nil {
		return err
	}

	return nil
}

// validateCustomRules applies custom validation rules
func (cv *ConfigValidator) validateCustomRules(config *Config) error {
	var errors []string

	// Database validation
	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		errors = append(errors, "database max_idle_conns cannot exceed max_open_conns")
	}

	// Redis validation
	if config.Redis.Enabled {
		if config.Redis.Host == "" {
			errors = append(errors, "redis host is required when redis is enabled")
		}
		if config.Redis.TTL < time.Minute {
			errors = append(errors, "redis TTL should be at least 1 minute")
		}
	}

	// JWT validation
	if config.JWT.AccessTokenTTL <= 0 {
		errors = append(errors, "JWT access token TTL must be positive")
	}
	if config.JWT.RefreshTokenTTL <= config.JWT.AccessTokenTTL {
		errors = append(errors, "JWT refresh token TTL must be greater than access token TTL")
	}

	// Server validation
	if config.Server.ReadTimeout < time.Second {
		errors = append(errors, "server read timeout should be at least 1 second")
	}
	if config.Server.WriteTimeout < time.Second {
		errors = append(errors, "server write timeout should be at least 1 second")
	}
	if config.Server.MaxRequestSize < 1024 {
		errors = append(errors, "server max request size should be at least 1KB")
	}

	// Payment validation
	if config.Payment.Provider != "mock" {
		if config.Payment.APIKey == "" {
			errors = append(errors, "payment API key is required for non-mock provider")
		}
		if config.Payment.SecretKey == "" {
			errors = append(errors, "payment secret key is required for non-mock provider")
		}
	}

	// Logging validation
	if config.Logging.Output == "file" && config.Logging.FilePath == "" {
		errors = append(errors, "logging file path is required when output is 'file'")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// formatValidationError formats validation errors for better readability
func (cv *ConfigValidator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []string
		for _, e := range validationErrors {
			errors = append(errors, cv.formatFieldError(e))
		}
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}
	return err
}

// formatFieldError formats a single field error
func (cv *ConfigValidator) formatFieldError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "port":
		return fmt.Sprintf("%s must be a valid port number (1-65535)", field)
	case "duration":
		return fmt.Sprintf("%s must be a valid duration", field)
	case "secret":
		return fmt.Sprintf("%s must be at least 32 characters for security", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}

// Custom validation functions

func validateURL(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	_, err := url.Parse(value)
	return err == nil
}

func validateEmail(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(value)
}

func validatePort(fl validator.FieldLevel) bool {
	value := fl.Field().Int()
	return value >= 1 && value <= 65535
}

func validateDuration(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	_, err := time.ParseDuration(value)
	return err == nil
}

func validatePath(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	// Basic path validation
	return !strings.Contains(value, "\x00")
}

func validateSecret(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Will be caught by required validation
	}
	return len(value) >= 32
}
