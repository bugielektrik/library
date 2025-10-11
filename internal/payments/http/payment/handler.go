package payment

import (
	"github.com/go-chi/chi/v5"

	"library-service/internal/adapters/http/handlers"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
)

// PaymentHandler handles HTTP requests for payments.
//
// ORGANIZATION:
// This file contains only the handler struct, constructor, and route definitions.
// Handler methods are split across multiple files by feature area:
//   - payment_initiate.go: Payment creation flows (InitiatePayment, PayWithSavedCard)
//   - payment_manage.go: Payment management (CancelPayment, RefundPayment)
//   - payment_query.go: Read operations (VerifyPayment, ListMemberPayments)
//   - payment_callback.go: External webhook handling (HandleCallback)
//
// RATIONALE:
// Splitting by feature area (creation, management, query, webhook) makes it easier to:
//   - Navigate to relevant code when working on specific features
//   - Review changes in pull requests (smaller, focused diffs)
//   - Test individual feature areas in isolation
//   - Understand payment flow without scrolling through 400+ lines
//
// Each split file is ~80-100 lines, focused on a single responsibility.
type PaymentHandler struct {
	handlers.BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewPaymentHandler creates a new payment handler.
func NewPaymentHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *PaymentHandler {
	return &PaymentHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for payment endpoints.
//
// ROUTE STRUCTURE:
//
//	POST   /initiate              → initiatePayment (payment_initiate.go)
//	POST   /pay-with-card         → payWithSavedCard (payment_initiate.go)
//	POST   /callback              → handleCallback (payment_callback.go) - PUBLIC endpoint for gateway
//	GET    /{id}                  → verifyPayment (payment_query.go)
//	POST   /{id}/cancel           → cancelPayment (payment_manage.go)
//	POST   /{id}/refund           → refundPayment (payment_manage.go)
//	GET    /member/{memberId}     → listMemberPayments (payment_query.go)
//
// SECURITY:
//   - All routes require authentication EXCEPT /callback
//   - /callback is called by payment gateway, validated by signature/token
//   - Refund operations check for admin role
func (h *PaymentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Payment initiation endpoints
	r.Post("/initiate", h.initiatePayment)
	r.Post("/pay-with-card", h.payWithSavedCard)

	// Webhook endpoint (public, called by payment gateway)
	r.Post("/callback", h.handleCallback)

	// Payment-specific operations
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.verifyPayment)
		r.Post("/cancel", h.cancelPayment)
		r.Post("/refund", h.refundPayment)
	})

	// Query endpoints
	r.Get("/member/{memberId}", h.listMemberPayments)

	return r
}
