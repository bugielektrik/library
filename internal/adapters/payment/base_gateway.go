package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"library-service/pkg/logutil"
)

// BaseGateway provides common functionality for payment gateways
type BaseGateway struct {
	client    *http.Client
	baseURL   string
	authToken string
	logger    *zap.Logger
	getAuthFn func(ctx context.Context) (string, error)
}

// NewBaseGateway creates a new base gateway
func NewBaseGateway(baseURL string, getAuthFn func(ctx context.Context) (string, error)) *BaseGateway {
	return &BaseGateway{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   baseURL,
		logger:    zap.L().Named("payment_gateway"),
		getAuthFn: getAuthFn,
	}
}

// ExecuteAuthenticatedRequest performs an authenticated HTTP request
func (g *BaseGateway) ExecuteAuthenticatedRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body interface{},
	result interface{},
) error {
	logger := logutil.FromContext(ctx).Named("gateway_request")

	// Get auth token
	token, err := g.getAuthFn(ctx)
	if err != nil {
		logger.Error("failed to get auth token", zap.Error(err))
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Build request URL
	url := g.baseURL + endpoint

	// Marshal request body if provided
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			logger.Error("failed to marshal request body", zap.Error(err))
			return fmt.Errorf("invalid request data: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		logger.Debug("request body", zap.String("body", string(jsonBody)))
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return fmt.Errorf("request creation failed: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	logger.Info("executing gateway request",
		zap.String("method", method),
		zap.String("url", url),
	)

	resp, err := g.client.Do(req)
	if err != nil {
		logger.Error("request failed", zap.Error(err))
		return fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response", zap.Error(err))
		return fmt.Errorf("failed to read gateway response: %w", err)
	}

	logger.Debug("gateway response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// Check response status
	if resp.StatusCode >= 400 {
		logger.Error("gateway returned error",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return g.parseErrorResponse(resp.StatusCode, respBody)
	}

	// Unmarshal response if result is provided
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			logger.Error("failed to unmarshal response", zap.Error(err))
			return fmt.Errorf("invalid gateway response format: %w", err)
		}
	}

	logger.Info("gateway request successful")
	return nil
}

// parseErrorResponse extracts error details from gateway response
func (g *BaseGateway) parseErrorResponse(statusCode int, body []byte) error {
	// Try to parse as JSON error
	var errorResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    string `json:"code"`
		Details string `json:"details"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil {
		// Successfully parsed JSON error
		if errorResp.Error != "" {
			return fmt.Errorf("gateway error (code %d): %s", statusCode, errorResp.Error)
		}
		if errorResp.Message != "" {
			return fmt.Errorf("gateway error (code %d): %s", statusCode, errorResp.Message)
		}
	}

	// Fallback to raw response
	if len(body) > 0 {
		return fmt.Errorf("gateway error (code %d): %s", statusCode, string(body))
	}

	// Generic error based on status code
	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("gateway authentication failed")
	case http.StatusForbidden:
		return fmt.Errorf("gateway access forbidden")
	case http.StatusNotFound:
		return fmt.Errorf("gateway resource not found")
	case http.StatusTooManyRequests:
		return fmt.Errorf("gateway rate limit exceeded")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("gateway service unavailable")
	default:
		if statusCode >= 500 {
			return fmt.Errorf("gateway server error (code %d)", statusCode)
		}
		return fmt.Errorf("gateway client error (code %d)", statusCode)
	}
}

// BuildURL constructs a URL with query parameters
func (g *BaseGateway) BuildURL(endpoint string, params map[string]string) string {
	url := g.baseURL + endpoint
	if len(params) == 0 {
		return url
	}

	first := true
	for key, value := range params {
		if first {
			url += "?"
			first = false
		} else {
			url += "&"
		}
		url += fmt.Sprintf("%s=%s", key, value)
	}

	return url
}

// LogGatewayOperation logs gateway operations with consistent format
func (g *BaseGateway) LogGatewayOperation(ctx context.Context, operation string, data interface{}) {
	logger := logutil.FromContext(ctx).Named("gateway")

	if jsonData, err := json.Marshal(data); err == nil {
		logger.Info(operation,
			zap.String("data", string(jsonData)),
			zap.Time("timestamp", time.Now()),
		)
	} else {
		logger.Info(operation,
			zap.Any("data", data),
			zap.Time("timestamp", time.Now()),
		)
	}
}

// HandleGatewayResponse processes common gateway response patterns
func (g *BaseGateway) HandleGatewayResponse(
	ctx context.Context,
	response interface{},
	expectedStatus string,
) error {
	// Common response structure
	type gatewayResponse struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	// Try to extract status from response
	if respBytes, err := json.Marshal(response); err == nil {
		var resp gatewayResponse
		if err := json.Unmarshal(respBytes, &resp); err == nil {
			if resp.Status != expectedStatus {
				return fmt.Errorf("unexpected gateway status: %s (expected %s), message: %s",
					resp.Status, expectedStatus, resp.Message)
			}
		}
	}

	return nil
}
