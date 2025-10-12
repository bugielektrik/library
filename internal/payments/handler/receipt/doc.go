// Package receipt provides HTTP handler for receipt generation and retrieval.
//
// This package handles receipt-related HTTP requests including:
//   - Generate receipt for payment (POST /receipts)
//   - Get receipt by ID (GET /receipts/{id})
//   - List member receipts (GET /receipts)
//
// Receipts are generated after successful payment completion and contain:
//   - Payment transaction details
//   - Member information
//   - Itemized breakdown
//   - Tax calculations
//   - Receipt number for accounting
//
// All endpoints require authentication (JWT middleware applied in router).
//
// Handler Organization:
//   - handler.go: Handler struct, routes, constructor, and all endpoint implementations
//
// Related Packages:
//   - Use Cases: internal/usecase/paymentops/ (receipt generation logic)
//   - Domain: internal/domain/payment/ (receipt entity)
//   - DTOs: internal/adapters/http/dto/receipt.go (request/response types)
//
// Example Usage:
//
//	receiptHandler := receipt.NewReceiptHandler(useCases, validator)
//	router.Group(func(r chi.Router) {
//	    r.Use(authMiddleware.Authenticate)
//	    r.Mount("/receipts", receiptHandler.Routes())
//	})
package receipt
