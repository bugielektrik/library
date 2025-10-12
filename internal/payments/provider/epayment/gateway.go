package epayment

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Timeout and expiry constants for payment provider configuration
const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests to the payment provider.
	// This includes OAuth token requests, payment status checks, refunds, and cancellations.
	DefaultHTTPTimeout = 30 * time.Second

	// MaxRetries is the maximum number of retry attempts for failed payment service.
	// Currently not implemented but reserved for future retry logic.
	MaxRetries = 3
)

// Config holds the configuration for edomain.kz provider.
type Config struct {
	// OAuth configuration
	ClientID     string // OAuth client ID
	ClientSecret string // OAuth client secret
	OAuthURL     string // OAuth token endpoint URL

	// Gateway configuration
	BaseURL     string // Base URL for payment API
	Terminal    string // Terminal ID for the merchant
	BackLink    string // URL where users are redirected after payment
	PostLink    string // URL where payment provider sends callbacks
	WidgetURL   string // URL for payment widget script
	Environment string // Environment (test/prod)
}

// Gateway represents the edomain.kz payment provider client.
//
// It provides methods for:
//   - OAuth authentication with automatic token caching
//   - Payment status checking
//   - Refunds (full and partial)
//   - Payment cancellation
//   - Saved card tokenization and charging
//
// The Gateway is thread-safe and manages OAuth tokens automatically,
// refreshing them before expiry to ensure uninterrupted operation.
type Gateway struct {
	config     *Config
	httpClient *http.Client
	logger     *zap.Logger

	// Token caching for OAuth
	token       string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// NewGateway creates a new payment provider instance.
//
// Parameters:
//   - config: Gateway configuration including OAuth credentials and API endpoints
//   - logger: Zap logger for structured logging
//
// Returns a fully initialized Gateway ready to process payments.
func NewGateway(config *Config, logger *zap.Logger) *Gateway {
	return &Gateway{
		config: config,
		httpClient: &http.Client{
			Timeout: DefaultHTTPTimeout,
		},
		logger: logger,
	}
}

// GetTerminal returns the terminal ID for the merchant.
func (g *Gateway) GetTerminal() string {
	return g.config.Terminal
}

// GetBackLink returns the URL where users are redirected after domain.
func (g *Gateway) GetBackLink() string {
	return g.config.BackLink
}

// GetPostLink returns the URL where payment provider sends callbacks.
func (g *Gateway) GetPostLink() string {
	return g.config.PostLink
}

// GetBaseURL returns the base URL of the payment provider API.
func (g *Gateway) GetBaseURL() string {
	return g.config.BaseURL
}

// GetEnvironment returns the current environment (test/prod).
func (g *Gateway) GetEnvironment() string {
	return g.config.Environment
}

// GetWidgetURL returns the URL for the payment widget.
//
// This URL is used to embed the payment widget script in your application.
func (g *Gateway) GetWidgetURL() string {
	return g.config.WidgetURL
}
