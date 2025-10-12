package epayment

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Test constants
const (
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testTerminal     = "test-terminal-123"
)

// newTestGateway creates a provider configured for testing
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

// ================================================================================
// Authentication & Token Management Tests
// ================================================================================

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
