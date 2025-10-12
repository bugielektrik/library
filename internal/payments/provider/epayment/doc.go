// Package epayment provides integration with the edomain.kz payment provider.
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
// The provider uses OAuth 2.0 client credentials flow with automatic token caching:
//
//	provider := edomain.NewGateway(config, logger)
//	token, err := provider.GetAuthToken(ctx)
//
// Tokens are cached and automatically refreshed 5 minutes before expiry to prevent
// race conditions and ensure continuous operation.
//
// # Payment Operations
//
// The provider supports the full payment lifecycle:
//
//  1. Payment initiation (creates invoice at provider)
//  2. Payment status checking (polls for completion)
//  3. Refunds (full or partial)
//  4. Cancellations (for pending payments)
//
// Example payment flow:
//
//	// Check payment status
//	status, err := provider.CheckPaymentStatus(ctx, invoiceID)
//	if err != nil {
//	    return err
//	}
//
//	// Process refund if needed
//	if shouldRefund {
//	    refund, err := provider.RefundPayment(ctx, invoiceID, amount)
//	}
//
// # Saved Cards
//
// The provider supports tokenization for recurring payments:
//
//	req := &domain.CardChargeRequest{
//		InvoiceID:   invoiceID,
//		Amount:      amount,
//		Currency:    "KZT",
//		CardID:      cardToken,
//		Description: "Payment description",
//	}
//	result, err := provider.ChargeCard(ctx, req)
//
// # Error Handling
//
// All provider methods return descriptive errors with context. Network errors,
// authentication failures, and provider-specific errors are properly wrapped
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
// The provider supports both real integration tests (requires test credentials)
// and mock testing via interface implementation.
package epayment
