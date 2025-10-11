package dto

// Payment callback constants for edomain.kz gateway integration.
//
// These constants define the standard codes and reasons used in payment
// gateway callbacks (PaymentCallbackRequest) to indicate transaction outcomes.

// Callback result codes sent by the payment gateway
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
	// Maps to domain domain.StatusCompleted.
	PaymentStatusSuccess = "success"

	// PaymentStatusFailed is the internal status for failed payments.
	// Maps to domain domain.StatusFailed.
	PaymentStatusFailed = "failed"
)

// IsSuccessfulCallback returns true if the callback indicates a successful domain.
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

// IsFailedCallback returns true if the callback indicates a failed domain.
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
// This helper replaces the inline status mapping logic in handlers.
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
