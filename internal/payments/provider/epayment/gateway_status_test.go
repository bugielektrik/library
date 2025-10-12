package epayment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test constants for status tests
const (
	testInvoiceID = "INV_TEST_001"
)

// ================================================================================
// Payment Status Check Tests
// ================================================================================

// TestCheckPaymentStatus_Success tests successful payment status check
func TestCheckPaymentStatus_Success(t *testing.T) {
	// Create mock servers
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", authHeader)
		}

		// Verify URL path
		if !strings.Contains(r.URL.Path, testInvoiceID) {
			t.Errorf("Expected path to contain invoice ID %s, got %s", testInvoiceID, r.URL.Path)
		}

		// Send success response
		resp := TransactionStatusResponse{
			ResultCode:    "0",
			ResultMessage: "Success",
			Transaction: TransactionDetails{
				ID:        "TX_123",
				InvoiceID: testInvoiceID,
				Amount:    10050,
				Currency:  "KZT",
				Status:    "COMPLETED",
				CardMask:  "4405-62**-****-1448",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	result, err := gateway.CheckPaymentStatus(context.Background(), testInvoiceID)
	if err != nil {
		t.Fatalf("CheckPaymentStatus failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.ResultCode != "0" {
		t.Errorf("Expected ResultCode '0', got '%s'", result.ResultCode)
	}

	if result.Transaction.InvoiceID != testInvoiceID {
		t.Errorf("Expected InvoiceID '%s', got '%s'", testInvoiceID, result.Transaction.InvoiceID)
	}
}

// TestCheckPaymentStatus_InvalidInvoice tests not found handling
func TestCheckPaymentStatus_InvalidInvoice(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]string{
			"error":   "NOT_FOUND",
			"message": "Invoice not found",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	// Should not return error for 404 - just malformed response
	_, err := gateway.CheckPaymentStatus(context.Background(), "INVALID_INVOICE")
	if err == nil {
		t.Error("Expected error for invalid invoice, got nil")
	}
}
