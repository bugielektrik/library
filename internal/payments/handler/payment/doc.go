// Package payment provides HTTP handler for payment processing service.
//
// This package handles comprehensive payment-related HTTP requests including:
//   - Initiate payment transaction (POST /payments/initiate)
//   - Pay with saved card (POST /payments/pay-with-card)
//   - Verify payment status (GET /payments/{id}/verify)
//   - List member payments (GET /payments)
//   - Cancel pending payment (POST /payments/{id}/cancel)
//   - Refund completed payment (POST /payments/{id}/refund)
//   - Handle provider callback (POST /payments/callback) - Public endpoint
//
// Handler Organization:
//   - handler.go: Handler struct, routes, and constructor
//   - initiate.go: Payment initiation flows (initiate, pay-with-card)
//   - manage.go: Payment management (cancel, refund)
//   - query.go: Read operations (verify, list)
//   - callback.go: External webhook handling from payment provider
//   - page.go: Payment widget page rendering
//
// This organization separates concerns by operation type:
//   - Initiation: Creating new payment transactions
//   - Management: Modifying existing payments (cancel, refund)
//   - Query: Reading payment status and history
//   - Callback: Asynchronous payment provider notifications
//
// Payment Flow:
//  1. Initiate: Member requests payment → Generate invoice with provider → Return payment URL
//  2. Process: Member completes payment on provider widget → Gateway sends callback
//  3. Callback: Verify signature → Update payment status → Trigger business logic
//  4. Verify: Poll payment status → Check with provider → Return current state
//
// Authentication:
//   - Most endpoints require JWT authentication
//   - Callback endpoint is public (validated by provider signature)
//
// Related Packages:
//   - Use Cases: internal/usecase/paymentops/ (payment business logic)
//   - Domain: internal/domain/payment/ (payment entity and service)
//   - DTOs: internal/adapters/http/dto/domain.go (request/response types)
//   - Gateway: internal/adapters/payment/epayment/ (ePayment provider integration)
//
// Example Usage:
//
//	paymentHandler := domain.NewPaymentHandler(useCases, validator)
//	router.Route("/payments", func(r chi.Router) {
//	    // Public callback endpoint
//	    r.Post("/callback", paymentHandler.HandleCallback)
//
//	    // Protected payment routes
//	    r.Group(func(r chi.Router) {
//	        r.Use(authMiddleware.Authenticate)
//	        r.Mount("/", paymentHandler.Routes())
//	    })
//	})
package payment
