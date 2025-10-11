package epayment

import (
	"fmt"
	"os"
)

// LoadConfigFromEnv loads epayment.kz configuration from environment variables.
func LoadConfigFromEnv() (*Config, error) {
	environment := getEnv("EPAYMENT_ENV", "test")

	config := &Config{
		ClientID:     os.Getenv("EPAYMENT_CLIENT_ID"),
		ClientSecret: os.Getenv("EPAYMENT_CLIENT_SECRET"),
		Terminal:     os.Getenv("EPAYMENT_TERMINAL"),
		BackLink:     os.Getenv("EPAYMENT_BACK_LINK"),
		PostLink:     os.Getenv("EPAYMENT_POST_LINK"),
		Environment:  environment,
	}

	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("EPAYMENT_CLIENT_ID environment variable is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("EPAYMENT_CLIENT_SECRET environment variable is required")
	}
	if config.Terminal == "" {
		return nil, fmt.Errorf("EPAYMENT_TERMINAL environment variable is required")
	}
	if config.BackLink == "" {
		return nil, fmt.Errorf("EPAYMENT_BACK_LINK environment variable is required")
	}
	if config.PostLink == "" {
		return nil, fmt.Errorf("EPAYMENT_POST_LINK environment variable is required")
	}

	// Set URLs based on environment
	if environment == "prod" {
		config.OAuthURL = "https://epay-oauth.homebank.kz/oauth2/token"
		config.BaseURL = "https://epay-api.homebank.kz"
		config.WidgetURL = "https://epay.homebank.kz/payform/payment-api.js"
	} else {
		config.OAuthURL = "https://test-epay-oauth.epayment.kz/oauth2/token"
		config.BaseURL = "https://test-epay-api.epayment.kz"
		config.WidgetURL = "https://test-epay.epayment.kz/redesign-payform/payment-api.js"
	}

	return config, nil
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
