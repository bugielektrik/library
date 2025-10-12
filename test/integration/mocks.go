//go:build integration
// +build integration

package integration

import (
	"context"
)

// MockPaymentGateway provides a test implementation of the payment provider
type MockPaymentGateway struct {
	terminal             string
	backLink             string
	postLink             string
	widgetURL            string
	checkPaymentResponse interface{}
}

func (m *MockPaymentGateway) GetAuthToken(ctx context.Context) (string, error) {
	return "test-token", nil
}

func (m *MockPaymentGateway) GetTerminal() string {
	return m.terminal
}

func (m *MockPaymentGateway) GetBackLink() string {
	return m.backLink
}

func (m *MockPaymentGateway) GetPostLink() string {
	return m.postLink
}

func (m *MockPaymentGateway) GetWidgetURL() string {
	return m.widgetURL
}

func (m *MockPaymentGateway) CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error) {
	if m.checkPaymentResponse != nil {
		return m.checkPaymentResponse, nil
	}
	return map[string]interface{}{
		"Status": "pending",
	}, nil
}

func (m *MockPaymentGateway) SetCheckPaymentResponse(resp interface{}) {
	m.checkPaymentResponse = resp
}

func stringPtr(s string) *string {
	return &s
}
