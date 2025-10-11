package epayment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"library-service/internal/domain/payment"
	"library-service/internal/infrastructure/log"
)

// CheckPaymentStatus checks the status of a payment transaction.
//
// API Endpoint: GET /check-status/payment/transaction/:invoiceid
//
// This method queries the epayment.kz gateway to retrieve the current status
// of a payment identified by its invoice ID. It returns detailed transaction
// information including payment status, amount, card details, and approval codes.
//
// Implements: payment.Gateway interface
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - invoiceID: Unique invoice identifier for the payment
//
// Returns transaction status details or an error if the request fails.
func (g *Gateway) CheckPaymentStatus(ctx context.Context, invoiceID string) (*payment.GatewayStatusResponse, error) {
	logger := log.FromContext(ctx).Named("check_payment_status").With(
		zap.String("invoice_id", invoiceID),
	)

	// Get auth token
	token, err := g.GetAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/check-status/payment/transaction/%s", g.config.BaseURL, invoiceID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status check request for invoice %s: %w", invoiceID, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send status check request", zap.Error(err))
		return nil, fmt.Errorf("failed to send status check request for invoice %s to %s: %w", invoiceID, url, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read status check response", zap.Error(err))
		return nil, fmt.Errorf("failed to read status check response for invoice %s: %w", invoiceID, err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Error("status check failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("status check failed for invoice %s with HTTP status %d: %s", invoiceID, resp.StatusCode, string(body))
	}

	// Parse response
	var statusResp TransactionStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		logger.Error("failed to parse status check response",
			zap.Error(err),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("failed to parse status check response for invoice %s: %w", invoiceID, err)
	}

	logger.Info("payment status checked",
		zap.String("result_code", statusResp.ResultCode),
		zap.String("result_message", statusResp.ResultMessage),
	)

	// Map epayment response to domain gateway response
	return &payment.GatewayStatusResponse{
		ResultCode:    statusResp.ResultCode,
		ResultMessage: statusResp.ResultMessage,
		Transaction: payment.GatewayTransactionDetails{
			ID:           statusResp.Transaction.ID,
			InvoiceID:    statusResp.Transaction.InvoiceID,
			Amount:       statusResp.Transaction.Amount,
			Currency:     statusResp.Transaction.Currency,
			Status:       statusResp.Transaction.Status,
			CardMask:     statusResp.Transaction.CardMask,
			ApprovalCode: statusResp.Transaction.ApprovalCode,
			Reference:    statusResp.Transaction.Reference,
		},
	}, nil
}

// RefundPayment processes a refund for a completed payment.
//
// API Endpoint: POST /operation/:id/refund
//
// This method initiates a refund for a completed transaction. It supports both
// full refunds (when amount is nil) and partial refunds (when amount is specified).
//
// Implements: payment.Gateway interface
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - transactionID: Gateway transaction identifier
//   - amount: Optional amount to refund (nil for full refund)
//   - externalID: Optional tracking ID for reconciliation
//
// Returns an error if the refund fails. The transaction must be in a completed
// state for the refund to succeed.
func (g *Gateway) RefundPayment(ctx context.Context, transactionID string, amount *float64, externalID string) error {
	logger := log.FromContext(ctx).Named("refund_payment").With(
		zap.String("transaction_id", transactionID),
	)

	// Get auth token
	token, err := g.GetAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build URL with optional query parameters
	url := fmt.Sprintf("%s/operation/%s/refund", g.config.BaseURL, transactionID)
	if amount != nil || externalID != "" {
		url += "?"
		if amount != nil {
			url += fmt.Sprintf("amount=%.2f", *amount)
		}
		if externalID != "" {
			if amount != nil {
				url += "&"
			}
			url += fmt.Sprintf("externalID=%s", externalID)
		}
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create refund request for transaction %s: %w", transactionID, err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("failed to send refund request", zap.Error(err))
		return fmt.Errorf("failed to send refund request for transaction %s to %s: %w", transactionID, url, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read refund response", zap.Error(err))
		return fmt.Errorf("failed to read refund response for transaction %s: %w", transactionID, err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		var refundResp RefundResponse
		if err := json.Unmarshal(body, &refundResp); err == nil {
			logger.Error("refund failed",
				zap.Int("status_code", resp.StatusCode),
				zap.Int("error_code", refundResp.Code),
				zap.String("message", refundResp.Message),
			)
			return fmt.Errorf("refund failed for transaction %s (code %d): %s", transactionID, refundResp.Code, refundResp.Message)
		}

		logger.Error("refund failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return fmt.Errorf("refund failed for transaction %s with HTTP status %d", transactionID, resp.StatusCode)
	}

	logger.Info("refund processed successfully")
	return nil
}

// CancelPayment cancels a payment in Auth status.
//
// API Endpoint: POST /operation/:id/cancel
//
// This method cancels a payment transaction that is in the authorization (Auth) state.
// Once a payment is completed, it cannot be cancelled - use RefundPayment instead.
//
// Implements: payment.Gateway interface
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - transactionID: Transaction ID to cancel
//
// Returns an error if the cancellation fails or the transaction is already completed.
func (g *Gateway) CancelPayment(ctx context.Context, transactionID string) error {
	logger := log.FromContext(ctx).Named("cancel_payment").With(
		zap.String("transaction_id", transactionID),
	)

	// Get auth token
	token, err := g.GetAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build URL
	url := fmt.Sprintf("%s/operation/%s/cancel", g.config.BaseURL, transactionID)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create cancel request for transaction %s: %w", transactionID, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send cancel request", zap.Error(err))
		return fmt.Errorf("failed to send cancel request for transaction %s to %s: %w", transactionID, url, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read cancel response", zap.Error(err))
		return fmt.Errorf("failed to read cancel response for transaction %s: %w", transactionID, err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		var cancelResp RefundResponse
		if err := json.Unmarshal(body, &cancelResp); err == nil {
			logger.Error("cancel failed",
				zap.Int("status_code", resp.StatusCode),
				zap.Int("error_code", cancelResp.Code),
				zap.String("message", cancelResp.Message),
			)
			return fmt.Errorf("cancel failed for transaction %s (code %d): %s", transactionID, cancelResp.Code, cancelResp.Message)
		}

		logger.Error("cancel failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return fmt.Errorf("cancel failed for transaction %s with HTTP status %d", transactionID, resp.StatusCode)
	}

	logger.Info("payment cancelled successfully")
	return nil
}

// ChargeCard charges a payment using a saved card token.
//
// API Endpoint: POST /payments/cards/auth
//
// This method initiates a payment using a previously saved card token.
// It's used for recurring payments or when the cardholder has authorized
// storing their card details for future transactions.
//
// Implements: payment.Gateway interface
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - req: Card payment request with amount, invoice ID, and saved card token
//
// Returns payment response with transaction details or an error if the charge fails.
func (g *Gateway) ChargeCard(ctx context.Context, req *payment.CardChargeRequest) (*payment.CardChargeResponse, error) {
	logger := log.FromContext(ctx).Named("charge_card_with_token").With(
		zap.String("invoice_id", req.InvoiceID),
		zap.String("card_id", req.CardID),
		zap.Int64("amount", req.Amount),
	)

	// Get auth token
	token, err := g.GetAuthToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build request body
	requestBody := map[string]interface{}{
		"amount":      req.Amount,
		"currency":    req.Currency,
		"terminalId":  g.config.Terminal,
		"invoiceId":   req.InvoiceID,
		"description": req.Description,
		"paymentType": "cardId",
		"cardId": map[string]string{
			"id": req.CardID,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("failed to marshal card payment request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal card payment request for invoice %s: %w", req.InvoiceID, err)
	}

	// Build URL
	url := fmt.Sprintf("%s/payments/cards/auth", g.config.BaseURL)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("failed to create card payment request", zap.Error(err))
		return nil, fmt.Errorf("failed to create card payment request for invoice %s: %w", req.InvoiceID, err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("failed to send card payment request", zap.Error(err))
		return nil, fmt.Errorf("failed to send card payment request for invoice %s to %s: %w", req.InvoiceID, url, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read card payment response", zap.Error(err))
		return nil, fmt.Errorf("failed to read card payment response for invoice %s: %w", req.InvoiceID, err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("card payment request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("card payment failed for invoice %s (card %s) with HTTP status %d: %s", req.InvoiceID, req.CardID, resp.StatusCode, string(body))
	}

	// Parse response
	var paymentResp CardPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		logger.Error("failed to parse card payment response", zap.Error(err))
		return nil, fmt.Errorf("failed to parse card payment response for invoice %s: %w", req.InvoiceID, err)
	}

	logger.Info("card payment processed successfully",
		zap.String("transaction_id", paymentResp.ID),
		zap.String("reference", paymentResp.Reference),
		zap.String("status", paymentResp.Status),
	)

	// Map epayment response to domain response
	return &payment.CardChargeResponse{
		ID:            paymentResp.ID,
		TransactionID: paymentResp.TransactionID,
		Status:        paymentResp.Status,
		Reference:     paymentResp.Reference,
		ApprovalCode:  paymentResp.ApprovalCode,
		ErrorCode:     paymentResp.ErrorCode,
		ErrorMessage:  paymentResp.ErrorMessage,
		ProcessedAt:   nil, // epayment doesn't provide this in the response
	}, nil
}

// ChargeCardWithToken is a deprecated alias for ChargeCard.
// Kept for backwards compatibility with existing code.
//
// Deprecated: Use ChargeCard instead.
func (g *Gateway) ChargeCardWithToken(ctx context.Context, req CardPaymentRequest) (*CardPaymentResponse, error) {
	// Convert to domain request
	domainReq := &payment.CardChargeRequest{
		InvoiceID:   req.InvoiceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CardID:      req.CardID,
		Description: req.Description,
	}

	// Call domain interface method
	resp, err := g.ChargeCard(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert back to legacy response
	return &CardPaymentResponse{
		ID:            resp.ID,
		TransactionID: resp.TransactionID,
		Status:        resp.Status,
		Reference:     resp.Reference,
		ApprovalCode:  resp.ApprovalCode,
		ErrorCode:     resp.ErrorCode,
		ErrorMessage:  resp.ErrorMessage,
	}, nil
}
