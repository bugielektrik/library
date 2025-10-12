// Package savedcard provides HTTP handler for saved payment card management.
//
// This package handles saved card-related HTTP requests including:
//   - Save payment card (POST /saved-cards)
//   - List member's saved cards (GET /saved-cards)
//   - Delete saved card (DELETE /saved-cards/{id})
//   - Set default card (PUT /saved-cards/{id}/default)
//
// Saved cards allow members to:
//   - Store payment methods securely via provider tokenization
//   - Make faster payments without re-entering card details
//   - Set a default card for automatic selection
//   - Manage multiple payment methods
//
// Security:
//   - Card data is NEVER stored in our database
//   - Only provider tokens (references) are stored
//   - PCI DSS compliance is provider's responsibility
//   - Members can only access their own saved cards
//
// All endpoints require authentication (JWT middleware applied in router).
//
// Handler Organization:
//   - handler.go: Handler struct, routes, and constructor
//   - crud.go: Create (Save), List, Delete operations
//   - manage.go: Card management (set default)
//
// Related Packages:
//   - Use Cases: internal/usecase/paymentops/ (saved card business logic)
//   - Domain: internal/domain/payment/ (saved card entity)
//   - DTOs: internal/adapters/http/dto/saved_card.go (request/response types)
//   - Gateway: internal/adapters/payment/epayment/ (card tokenization)
//
// Example Usage:
//
//	savedCardHandler := savedcard.NewSavedCardHandler(useCases, validator)
//	router.Group(func(r chi.Router) {
//	    r.Use(authMiddleware.Authenticate)
//	    r.Mount("/saved-cards", savedCardHandler.Routes())
//	})
package savedcard
