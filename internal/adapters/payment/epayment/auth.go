package epayment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"library-service/internal/infrastructure/log"
)

// TokenExpiryBuffer is subtracted from token expiry to avoid race conditions.
// We refresh tokens 5 minutes before they actually expire.
const TokenExpiryBuffer = 5 * time.Minute

// GetAuthToken retrieves a cached OAuth token or fetches a new one from epayment.kz.
//
// This method implements thread-safe token caching with automatic refresh:
//  1. Check cached token with read lock (fast path)
//  2. If expired, acquire write lock
//  3. Double-check after acquiring write lock (another goroutine might have refreshed)
//  4. Fetch new token from OAuth endpoint
//  5. Cache with 5-minute buffer before actual expiry
//
// The token is cached in memory and reused across all payment operations
// until it expires, minimizing API calls to the payment gateway.
func (g *Gateway) GetAuthToken(ctx context.Context) (string, error) {
	// Check if we have a valid cached token (fast path with read lock)
	g.tokenMutex.RLock()
	if g.token != "" && time.Now().Before(g.tokenExpiry) {
		token := g.token
		g.tokenMutex.RUnlock()
		return token, nil
	}
	g.tokenMutex.RUnlock()

	// Acquire write lock to fetch new token
	g.tokenMutex.Lock()
	defer g.tokenMutex.Unlock()

	// Double-check in case another goroutine already fetched the token
	if g.token != "" && time.Now().Before(g.tokenExpiry) {
		return g.token, nil
	}

	logger := log.FromContext(ctx).Named("get_auth_token")
	logger.Info("fetching new auth token from epayment.kz")

	// Prepare OAuth request with correct scope from epayment.kz docs
	data := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     g.config.ClientID,
		"client_secret": g.config.ClientSecret,
		"scope":         "webapi usermanagement email_send verification statement statistics payment",
		"terminal":      g.config.Terminal,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshal OAuth request", zap.Error(err))
		return "", fmt.Errorf("failed to marshal OAuth request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", g.config.OAuthURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create OAuth request", zap.Error(err))
		return "", fmt.Errorf("failed to create OAuth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send OAuth request", zap.Error(err))
		return "", fmt.Errorf("failed to send OAuth request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read OAuth response", zap.Error(err))
		return "", fmt.Errorf("failed to read OAuth response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("OAuth request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("OAuth request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		logger.Error("failed to parse OAuth response", zap.Error(err))
		return "", fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	// Cache token with expiry buffer to avoid using expired tokens due to clock skew
	g.token = tokenResp.AccessToken
	g.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - TokenExpiryBuffer)

	logger.Info("auth token fetched successfully",
		zap.String("token_type", tokenResp.TokenType),
		zap.Int("expires_in", tokenResp.ExpiresIn),
	)

	return g.token, nil
}
