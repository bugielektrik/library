package epayment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test constants for refund tests
const (
	testTransactionID = "TX_TEST_123"
)

// ================================================================================
// Refund & Cancel Operation Tests
// ================================================================================

// TestRefundPayment_FullRefund tests full refund processing
func TestRefundPayment_FullRefund(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify URL contains transaction ID
		if !strings.Contains(r.URL.Path, testTransactionID) {
			t.Errorf("Expected path to contain transaction ID %s, got %s", testTransactionID, r.URL.Path)
		}

		// Send success response
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	err := gateway.RefundPayment(context.Background(), testTransactionID, nil, "")
	if err != nil {
		t.Fatalf("RefundPayment failed: %v", err)
	}
}

// TestRefundPayment_PartialRefund tests partial refund processing
func TestRefundPayment_PartialRefund(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters for partial refund
		amountParam := r.URL.Query().Get("amount")
		if amountParam == "" {
			t.Error("Expected amount parameter for partial refund")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	partialAmount := 50.00

	err := gateway.RefundPayment(context.Background(), testTransactionID, &partialAmount, "REFUND_001")
	if err != nil {
		t.Fatalf("RefundPayment (partial) failed: %v", err)
	}
}

// TestRefundPayment_GatewayError tests provider error handling
func TestRefundPayment_GatewayError(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.WriteHeader(http.StatusBadRequest)
		resp := RefundResponse{
			Code:    400,
			Message: "Transaction already refunded",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	err := gateway.RefundPayment(context.Background(), testTransactionID, nil, "")
	if err == nil {
		t.Error("Expected error for provider error response, got nil")
	}

	if !strings.Contains(err.Error(), "Transaction already refunded") {
		t.Errorf("Expected error message to contain 'Transaction already refunded', got: %v", err)
	}
}

// TestCancelPayment_Success tests successful payment cancellation
func TestCancelPayment_Success(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify path contains "cancel"
		if !strings.Contains(r.URL.Path, "cancel") {
			t.Errorf("Expected path to contain 'cancel', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	err := gateway.CancelPayment(context.Background(), testTransactionID)
	if err != nil {
		t.Fatalf("CancelPayment failed: %v", err)
	}
}

// TestCancelPayment_AlreadyCompleted tests cancellation of completed payment
func TestCancelPayment_AlreadyCompleted(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error for already completed transaction
		w.WriteHeader(http.StatusBadRequest)
		resp := RefundResponse{
			Code:    400,
			Message: "Cannot cancel completed transaction",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	err := gateway.CancelPayment(context.Background(), testTransactionID)
	if err == nil {
		t.Error("Expected error for cancelling completed payment, got nil")
	}
}
