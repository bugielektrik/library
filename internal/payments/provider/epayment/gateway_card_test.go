package epayment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"library-service/internal/payments/domain"
)

// ================================================================================
// Card Charging Tests
// ================================================================================

// TestChargeCard_Success tests saved card charging
func TestChargeCard_Success(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and content type
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]interface{}
		json.Unmarshal(body, &reqData)

		if reqData["paymentType"] != "cardId" {
			t.Errorf("Expected paymentType: cardId, got %v", reqData["paymentType"])
		}

		// Send success response
		resp := CardPaymentResponse{
			ID:        "TX_CARD_123",
			Reference: "REF_123",
			Status:    "COMPLETED",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	req := &domain.CardChargeRequest{
		Amount:      10000,
		Currency:    "KZT",
		InvoiceID:   "INV_TEST_001",
		Description: "Test payment",
		CardID:      "CARD_TOKEN_456",
	}

	result, err := gateway.ChargeCard(context.Background(), req)
	if err != nil {
		t.Fatalf("ChargeCard failed: %v", err)
	}

	if result.ID != "TX_CARD_123" {
		t.Errorf("Expected transaction ID 'TX_CARD_123', got '%s'", result.ID)
	}

	if result.Status != "COMPLETED" {
		t.Errorf("Expected status 'COMPLETED', got '%s'", result.Status)
	}
}

// TestChargeCard_InvalidCard tests invalid card handling
func TestChargeCard_InvalidCard(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error for invalid card
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"INVALID_CARD","message":"Card not found or expired"}`))
	}))
	defer apiServer.Close()

	gateway := newTestGateway(apiServer.URL, oauthServer.URL)

	req := &domain.CardChargeRequest{
		Amount:    10000,
		Currency:  "KZT",
		InvoiceID: "INV_TEST_001",
		CardID:    "INVALID_CARD",
	}

	_, err := gateway.ChargeCard(context.Background(), req)
	if err == nil {
		t.Error("Expected error for invalid card, got nil")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error to mention HTTP 400, got: %v", err)
	}
}
