package payment

import (
	"library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/service/payment"
)

// ============================================================================
// Payment Callback DTOs
// ============================================================================

// PaymentCallbackRequest represents the asynchronous callback request from the payment provider.
//
// This DTO matches the exact format sent by epayment.kz provider when a payment status changes.
// The provider sends this webhook to the PostLink URL configured during payment initiation.
//
// The callback is used to update payment status asynchronously after the user completes
// the payment flow in the provider's widget.
//
// Fields:
//   - Code: Result code - "ok" for success, "error" for failure
//   - InvoiceID: Invoice/order identifier matching the one from InitiatePaymentResponse
//   - Amount: Transaction amount in smallest currency unit
//   - Currency: ISO 4217 currency code
//   - CardMask: Masked card number (e.g., "4405-62**-****-1448") if card payment
//   - Reason: Status reason - "success" for successful payment or error description
//   - ReasonCode: Numeric error code if payment failed
//   - TransactionID: Gateway-assigned transaction identifier for tracking
//   - Reference: Additional reference number from provider
//   - ApprovalCode: Bank approval code for successful card transactions
//   - Terminal: Terminal ID where transaction was processed
//   - Extra: Additional provider-specific fields not in standard format
//
// Security Note: In production, callback requests should be validated using
// signature verification or IP whitelisting to prevent fraudulent callbacks.
//
// Example successful callback JSON:
//
//	{
//	  "code": "ok",
//	  "invoiceId": "INV_123456",
//	  "amount": 5000,
//	  "currency": "KZT",
//	  "cardMask": "4405-62**-****-1448",
//	  "reason": "success",
//	  "transactionId": "GW_TX_789",
//	  "approvalCode": "ABC123"
//	}
type PaymentCallbackRequest struct {
	Code          string                 `json:"code"`                    // "ok" or "error"
	InvoiceID     string                 `json:"invoiceId"`               // Order number
	Amount        int64                  `json:"amount"`                  // Transaction amount
	Currency      string                 `json:"currency"`                // Currency code
	CardMask      *string                `json:"cardMask,omitempty"`      // Masked card number
	Reason        string                 `json:"reason"`                  // "success" or error description
	ReasonCode    *string                `json:"reasonCode,omitempty"`    // Error numeric code
	TransactionID *string                `json:"transactionId,omitempty"` // Gateway transaction ID
	Reference     *string                `json:"reference,omitempty"`     // Reference number
	ApprovalCode  *string                `json:"approvalCode,omitempty"`  // Approval code
	Terminal      *string                `json:"terminal,omitempty"`      // Terminal ID
	Extra         map[string]interface{} `json:"extra,omitempty"`         // Additional fields
}

// PaymentCallbackResponse represents the acknowledgement response sent back to the payment provider.
//
// When the provider sends a callback to our webhook endpoint, we process the payment
// status update and respond with this DTO to acknowledge receipt and processing.
//
// Fields:
//   - PaymentID: Our internal payment identifier
//   - Status: Updated payment status after processing the callback
//   - Message: Human-readable confirmation message
//
// The provider may retry callbacks if it doesn't receive a successful HTTP 200 response
// with this payload within a reasonable timeout.
type PaymentCallbackResponse struct {
	PaymentID string        `json:"payment_id"`
	Status    domain.Status `json:"status"`
	Message   string        `json:"message"`
}

// ToPaymentCallbackResponse converts a use case HandleCallbackResponse to DTO format.
//
// This helper creates the acknowledgement response sent back to the payment provider
// after processing a webhook callback. It includes a hardcoded success message.
//
// Used by: POST /payments/callback handler (webhook endpoint)
func ToPaymentCallbackResponse(resp paymentops.HandleCallbackResponse) PaymentCallbackResponse {
	return PaymentCallbackResponse{
		PaymentID: resp.PaymentID,
		Status:    resp.Status,
		Message:   "Payment callback processed successfully",
	}
}

// ============================================================================
// Payment Callback Constants
// ============================================================================

// Payment callback constants for epayment.kz provider integration.
//
// These constants define the standard codes and reasons used in payment
// provider callbacks (PaymentCallbackRequest) to indicate transaction outcomes.

// Callback result codes sent by the payment provider
const (
	// CallbackCodeOK indicates the payment transaction was successful.
	// This code is sent when the payment was authorized and completed successfully.
	CallbackCodeOK = "ok"

	// CallbackCodeError indicates the payment transaction failed.
	// This code is sent when the payment was declined, cancelled, or encountered an error.
	CallbackCodeError = "error"
)

// Callback reason strings for successful and failed transactions
const (
	// CallbackReasonSuccess indicates a successful payment completion.
	// Used in conjunction with CallbackCodeOK to confirm successful authorization.
	CallbackReasonSuccess = "success"

	// CallbackReasonFailed is a generic failure reason.
	// The actual failure reason may be more specific and provided in the Reason field.
	CallbackReasonFailed = "failed"

	// CallbackReasonDeclined indicates the payment was declined by the issuing bank.
	// Common reasons include insufficient funds, card restrictions, or fraud detection.
	CallbackReasonDeclined = "declined"

	// CallbackReasonCancelled indicates the payment was cancelled by the user.
	// This occurs when the user explicitly cancels the payment flow in the widget.
	CallbackReasonCancelled = "cancelled"

	// CallbackReasonTimeout indicates the payment session expired.
	// Payments typically have a 15-30 minute window for completion.
	CallbackReasonTimeout = "timeout"

	// CallbackReasonInvalidCard indicates the card details were invalid.
	// This includes wrong card number, expired card, or invalid CVV.
	CallbackReasonInvalidCard = "invalid_card"
)

// Payment status mapping constants
const (
	// PaymentStatusSuccess is the internal status for successful payments.
	// Maps to domain.StatusCompleted.
	PaymentStatusSuccess = "success"

	// PaymentStatusFailed is the internal status for failed payments.
	// Maps to domain.StatusFailed.
	PaymentStatusFailed = "failed"
)

// ============================================================================
// Callback Helper Functions
// ============================================================================

// IsSuccessfulCallback returns true if the callback indicates a successful payment.
//
// A callback is considered successful when both:
//   - Code is "ok"
//   - Reason is "success"
//
// Example usage:
//
//	if IsSuccessfulCallback(req.Code, req.Reason) {
//	    // Process successful payment
//	}
func IsSuccessfulCallback(code, reason string) bool {
	return code == CallbackCodeOK && reason == CallbackReasonSuccess
}

// IsFailedCallback returns true if the callback indicates a failed payment.
//
// A callback is considered failed when:
//   - Code is "error", OR
//   - Reason is not "success"
//
// Example usage:
//
//	if IsFailedCallback(req.Code, req.Reason) {
//	    // Handle payment failure
//	}
func IsFailedCallback(code, reason string) bool {
	return code == CallbackCodeError || reason != CallbackReasonSuccess
}

// GetPaymentStatus converts callback code and reason to internal payment status.
//
// Returns:
//   - "success" if code is "ok" and reason is "success"
//   - "failed" otherwise
//
// This helper replaces the inline status mapping logic in handler.
//
// Example usage:
//
//	status := GetPaymentStatus(req.Code, req.Reason)
func GetPaymentStatus(code, reason string) string {
	if IsSuccessfulCallback(code, reason) {
		return PaymentStatusSuccess
	}
	return PaymentStatusFailed
}
