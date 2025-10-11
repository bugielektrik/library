package epayment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Test constants
const (
	testClientID      = "test-client-id"
	testClientSecret  = "test-client-secret"
	testTerminal      = "test-terminal-123"
	testInvoiceID     = "INV_TEST_001"
	testTransactionID = "TX_TEST_123"
)

// newTestGateway creates a gateway configured for testing
func newTestGateway(baseURL, oauthURL string) *Gateway {
	config := &Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Terminal:     testTerminal,
		BaseURL:      baseURL,
		OAuthURL:     oauthURL,
		WidgetURL:    "https://test-widget.edomain.kz",
		BackLink:     "https://test.example.com/back",
		PostLink:     "https://test.example.com/callback",
		Environment:  "test",
	}

	logger := zap.NewNop() // Use no-op logger for tests
	return NewGateway(config, logger)
}

// TestGetAuthToken_Success tests successful OAuth token retrieval
func TestGetAuthToken_Success(t *testing.T) {
	// Create mock OAuth server
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var reqData map[string]string
		json.Unmarshal(body, &reqData)

		if reqData["grant_type"] != "client_credentials" {
			t.Errorf("Expected grant_type: client_credentials, got %s", reqData["grant_type"])
		}

		// Send success response
		resp := TokenResponse{
			AccessToken: "test-access-token-12345",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "payment",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	gateway := newTestGateway("https://test-api.edomain.kz", oauthServer.URL)

	token, err := gateway.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("GetAuthToken failed: %v", err)
	}

	if token != "test-access-token-12345" {
		t.Errorf("Expected token 'test-access-token-12345', got '%s'", token)
	}
}

// TestGetAuthToken_CacheHit tests that token caching works
func TestGetAuthToken_CacheHit(t *testing.T) {
	requestCount := 0

	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := TokenResponse{
			AccessToken: "cached-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			Scope:       "payment",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	gateway := newTestGateway("https://test-api.edomain.kz", oauthServer.URL)

	// First call - should fetch new token
	token1, err := gateway.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("First GetAuthToken failed: %v", err)
	}

	// Second call - should use cached token
	token2, err := gateway.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetAuthToken failed: %v", err)
	}

	if token1 != token2 {
		t.Errorf("Expected same token from cache, got different tokens")
	}

	if requestCount != 1 {
		t.Errorf("Expected 1 OAuth request (cached on second call), got %d", requestCount)
	}
}

// TestGetAuthToken_CacheExpiry tests token refresh before expiry
func TestGetAuthToken_CacheExpiry(t *testing.T) {
	requestCount := 0

	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		resp := TokenResponse{
			AccessToken: "short-lived-token",
			TokenType:   "Bearer",
			ExpiresIn:   1, // 1 second expiry
			Scope:       "payment",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	gateway := newTestGateway("https://test-api.edomain.kz", oauthServer.URL)

	// First call
	_, err := gateway.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("First GetAuthToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Second call - should fetch new token
	_, err = gateway.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetAuthToken failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("Expected 2 OAuth requests (token expired), got %d", requestCount)
	}
}

// TestGetAuthToken_NetworkError tests network failure handling
func TestGetAuthToken_NetworkError(t *testing.T) {
	// Create server that immediately closes connections
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force connection close
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	serverURL := oauthServer.URL
	oauthServer.Close() // Close server to simulate network error

	gateway := newTestGateway("https://test-api.edomain.kz", serverURL)

	_, err := gateway.GetAuthToken(context.Background())
	if err == nil {
		t.Error("Expected error for network failure, got nil")
	}
}

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

// TestRefundPayment_GatewayError tests gateway error handling
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
		t.Error("Expected error for gateway error response, got nil")
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

// TestChargeCardWithToken_Success tests saved card charging
func TestChargeCardWithToken_Success(t *testing.T) {
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

	req := CardPaymentRequest{
		Amount:      10000,
		Currency:    "KZT",
		InvoiceID:   testInvoiceID,
		Description: "Test payment",
		CardID:      "CARD_TOKEN_456",
	}

	result, err := gateway.ChargeCardWithToken(context.Background(), req)
	if err != nil {
		t.Fatalf("ChargeCardWithToken failed: %v", err)
	}

	if result.ID != "TX_CARD_123" {
		t.Errorf("Expected transaction ID 'TX_CARD_123', got '%s'", result.ID)
	}

	if result.Status != "COMPLETED" {
		t.Errorf("Expected status 'COMPLETED', got '%s'", result.Status)
	}
}

// TestChargeCardWithToken_InvalidCard tests invalid card handling
func TestChargeCardWithToken_InvalidCard(t *testing.T) {
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

	req := CardPaymentRequest{
		Amount:    10000,
		Currency:  "KZT",
		InvoiceID: testInvoiceID,
		CardID:    "INVALID_CARD",
	}

	_, err := gateway.ChargeCardWithToken(context.Background(), req)
	if err == nil {
		t.Error("Expected error for invalid card, got nil")
	}

	if !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error to mention HTTP 400, got: %v", err)
	}
}

// TestConcurrentTokenRequests tests thread safety of token caching
func TestConcurrentTokenRequests(t *testing.T) {
	requestCount := 0
	var requestMutex sync.Mutex

	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		requestCount++
		requestMutex.Unlock()

		// Simulate slow token generation
		time.Sleep(50 * time.Millisecond)

		resp := TokenResponse{AccessToken: "concurrent-token", TokenType: "Bearer", ExpiresIn: 3600}
		json.NewEncoder(w).Encode(resp)
	}))
	defer oauthServer.Close()

	gateway := newTestGateway("https://test-api.edomain.kz", oauthServer.URL)

	// Launch 10 concurrent token requests
	const goroutines = 10
	errors := make(chan error, goroutines)
	tokens := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			token, err := gateway.GetAuthToken(context.Background())
			errors <- err
			tokens <- token
		}()
	}

	// Collect results
	for i := 0; i < goroutines; i++ {
		if err := <-errors; err != nil {
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
		<-tokens
	}

	// Should only make 1 OAuth request despite 10 concurrent calls
	if requestCount != 1 {
		t.Errorf("Expected 1 OAuth request (others should use cache), got %d", requestCount)
	}
}
