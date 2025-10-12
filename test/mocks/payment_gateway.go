//go:build integration
// +build integration

package mocks

import (
	"context"
)

// PaymentGateway provides a test implementation of the payment provider interface.
// This mock is used in integration tests to simulate payment provider behavior
// without making actual HTTP requests to external service.
type PaymentGateway struct {
	terminal             string
	backLink             string
	postLink             string
	widgetURL            string
	checkPaymentResponse interface{}
}

// NewPaymentGateway creates a new mock payment provider with default values.
func NewPaymentGateway() *PaymentGateway {
	return &PaymentGateway{
		terminal:  "test-terminal",
		backLink:  "http://localhost:8080/payment/callback",
		postLink:  "http://localhost:8080/payment/webhook",
		widgetURL: "http://test-widget.example.com",
	}
}

// GetAuthToken returns a test authentication token.
func (m *PaymentGateway) GetAuthToken(ctx context.Context) (string, error) {
	return "test-token", nil
}

// GetTerminal returns the configured terminal ID.
func (m *PaymentGateway) GetTerminal() string {
	return m.terminal
}

// GetBackLink returns the configured back link URL.
func (m *PaymentGateway) GetBackLink() string {
	return m.backLink
}

// GetPostLink returns the configured post link URL.
func (m *PaymentGateway) GetPostLink() string {
	return m.postLink
}

// GetWidgetURL returns the configured widget URL.
func (m *PaymentGateway) GetWidgetURL() string {
	return m.widgetURL
}

// CheckPaymentStatus checks the payment status for a given invoice ID.
// Returns the configured response or a default pending status.
func (m *PaymentGateway) CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error) {
	if m.checkPaymentResponse != nil {
		return m.checkPaymentResponse, nil
	}
	return map[string]interface{}{
		"Status": "pending",
	}, nil
}

// SetCheckPaymentResponse sets a custom response for CheckPaymentStatus.
// This allows tests to simulate different payment statuses.
func (m *PaymentGateway) SetCheckPaymentResponse(resp interface{}) {
	m.checkPaymentResponse = resp
}

// SetTerminal sets the terminal ID for testing.
func (m *PaymentGateway) SetTerminal(terminal string) {
	m.terminal = terminal
}

// SetBackLink sets the back link URL for testing.
func (m *PaymentGateway) SetBackLink(backLink string) {
	m.backLink = backLink
}

// SetPostLink sets the post link URL for testing.
func (m *PaymentGateway) SetPostLink(postLink string) {
	m.postLink = postLink
}

// SetWidgetURL sets the widget URL for testing.
func (m *PaymentGateway) SetWidgetURL(widgetURL string) {
	m.widgetURL = widgetURL
}
