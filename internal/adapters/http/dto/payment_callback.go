package dto

import (
	"library-service/internal/domain/payment"
	"library-service/internal/usecase/paymentops"
)

// PaymentCallbackRequest represents the asynchronous callback request from the payment gateway.
//
// This DTO matches the exact format sent by epayment.kz gateway when a payment status changes.
// The gateway sends this webhook to the PostLink URL configured during payment initiation.
//
// The callback is used to update payment status asynchronously after the user completes
// the payment flow in the gateway's widget.
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
//   - Reference: Additional reference number from gateway
//   - ApprovalCode: Bank approval code for successful card transactions
//   - Terminal: Terminal ID where transaction was processed
//   - Extra: Additional gateway-specific fields not in standard format
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

// PaymentCallbackResponse represents the acknowledgement response sent back to the payment gateway.
//
// When the gateway sends a callback to our webhook endpoint, we process the payment
// status update and respond with this DTO to acknowledge receipt and processing.
//
// Fields:
//   - PaymentID: Our internal payment identifier
//   - Status: Updated payment status after processing the callback
//   - Message: Human-readable confirmation message
//
// The gateway may retry callbacks if it doesn't receive a successful HTTP 200 response
// with this payload within a reasonable timeout.
type PaymentCallbackResponse struct {
	PaymentID string         `json:"payment_id"`
	Status    payment.Status `json:"status"`
	Message   string         `json:"message"`
}

// ToPaymentCallbackResponse converts a use case HandleCallbackResponse to DTO format.
//
// This helper creates the acknowledgement response sent back to the payment gateway
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
