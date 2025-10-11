// Package payment provides payment gateway adapter implementations.
//
// This package implements outbound adapters for external payment processing
// services. Currently supports epayment.kz (Kazakhstan payment gateway) with
// extensible design for adding additional gateways.
//
// Implementations:
//   - epayment/: epayment.kz gateway adapter (primary for KZ market)
//
// Payment gateway interface:
//
//	type Gateway interface {
//	    InitiatePayment(ctx context.Context, req PaymentRequest) (PaymentResponse, error)
//	    VerifyCallback(ctx context.Context, callback CallbackData) (VerificationResult, error)
//	    RefundPayment(ctx context.Context, paymentID string, amount int) error
//	    GetPaymentStatus(ctx context.Context, paymentID string) (PaymentStatus, error)
//	}
//
// epayment.kz integration:
//   - Payment methods: Card, QR code, mobile wallets
//   - Currency: KZT (Kazakhstani Tenge)
//   - Callback verification: SHA-256 HMAC signature
//   - Webhook endpoint: /api/v1/payments/callback
//
// Payment flow:
//  1. InitiatePayment: Create payment session, get redirect URL
//  2. User redirected to payment gateway
//  3. User completes payment on gateway site
//  4. Gateway sends callback to webhook with signature
//  5. VerifyCallback: Validate signature, update payment status
//  6. Generate receipt for successful payments
//
// Security:
//   - Webhook signatures verified with secret key
//   - Payment amounts in smallest currency unit (tiyns for KZT)
//   - HTTPS required for all gateway communication
//   - Secret keys stored in environment variables
//   - Idempotency keys prevent duplicate payments
//
// Error handling:
//   - Gateway timeout: 30 seconds with retries
//   - Network errors: Exponential backoff retry (max 3 attempts)
//   - Invalid signatures: Payment rejected, logged for investigation
//   - Insufficient funds: User notified, payment marked failed
//
// Testing:
//   - Sandbox/test mode for development
//   - Mock gateway for unit tests
//   - Test card numbers provided by gateway
//
// Configuration:
//   - Gateway credentials via environment (EPAYMENT_MERCHANT_ID, EPAYMENT_SECRET)
//   - Webhook URL configuration
//   - Sandbox mode toggle
//   - Timeout and retry settings
//
// Future gateways:
//   - Implement Gateway interface for new provider
//   - Add to adapters/payment/ directory
//   - Configure in container.go via environment flag
package payment
