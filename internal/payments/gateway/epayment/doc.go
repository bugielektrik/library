// Package epayment provides integration with the edomain.kz payment gateway.
//
// This package implements the PaymentGateway interface for Kazakhstan's edomain.kz
// payment processing service, providing a complete payment solution including OAuth
// authentication, payment processing, refunds, cancellations, and saved card tokenization.
//
// # Architecture
//
// The Gateway type is the main entry point that handles all interactions with the
// edomain.kz API. It implements automatic token management with caching and refresh,
// ensuring optimal performance and reliability.
//
// # Authentication
//
// The gateway uses OAuth 2.0 client credentials flow with automatic token caching:
//
//	gateway := edomain.NewGateway(config, logger)
//	token, err := gateway.GetAuthToken(ctx)
//
// Tokens are cached and automatically refreshed 5 minutes before expiry to prevent
// race conditions and ensure continuous operation.
//
// # Payment Operations
//
// The gateway supports the full payment lifecycle:
//
//  1. Payment initiation (creates invoice at gateway)
//  2. Payment status checking (polls for completion)
//  3. Refunds (full or partial)
//  4. Cancellations (for pending payments)
//
// Example payment flow:
//
//	// Check payment status
//	status, err := gateway.CheckPaymentStatus(ctx, invoiceID)
//	if err != nil {
//	    return err
//	}
//
//	// Process refund if needed
//	if shouldRefund {
//	    refund, err := gateway.RefundPayment(ctx, invoiceID, amount)
//	}
//
// # Saved Cards
//
// The gateway supports tokenization for recurring payments:
//
//	result, err := gateway.ChargeCardWithToken(ctx, cardToken, amount, invoiceID)
//
// # Error Handling
//
// All gateway methods return descriptive errors with context. Network errors,
// authentication failures, and gateway-specific errors are properly wrapped
// with the %w verb for error unwrapping.
//
// # Configuration
//
// The Config struct requires:
//   - ClientID and ClientSecret: OAuth credentials from edomain.kz
//   - Terminal: Merchant terminal ID
//   - BaseURL: API endpoint (different for test/production)
//   - OAuthURL: OAuth token endpoint
//   - WidgetURL: JavaScript widget URL for frontend integration
//   - BackLink: URL where users are redirected after payment
//   - PostLink: Webhook URL for payment status callbacks
//   - Environment: "test" or "prod"
//
// # Thread Safety
//
// The Gateway type is safe for concurrent use. Token caching uses read-write
// locks to allow concurrent reads while preventing race conditions during
// token refresh.
//
// # Testing
//
// For testing, use the test environment configuration provided by edomain.kz.
// The gateway supports both real integration tests (requires test credentials)
// and mock testing via interface implementation.
package epayment
