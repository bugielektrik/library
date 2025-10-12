// Package savedcard implements use cases for saved payment card management.
//
// This package orchestrates workflows for storing, managing, and using saved
// payment cards for future transactions. It enables members to save card tokens
// for one-click payments while maintaining PCI DSS compliance by never storing
// actual card numbers.
//
// Use cases implemented:
//   - ListSavedCardsUseCase: Retrieves member's saved cards
//   - DeleteSavedCardUseCase: Removes a saved card
//   - PayWithSavedCardUseCase: Initiates payment using stored card token
//
// Dependencies:
//   - domain.SavedCardRepository: For card token persistence
//   - domain.Repository: For payment record creation
//   - domain.Service: For payment validation
//   - domain.Gateway: External payment provider for card charging
//
// Example usage:
//
//	listUC := savedcard.NewListSavedCardsUseCase(savedCardRepo)
//	response, err := listUC.Execute(ctx, savedcard.ListSavedCardsRequest{
//	    MemberID: "member-uuid",
//	})
//	// response contains: List of saved cards with masked numbers
//
//	payUC := savedcard.NewPayWithSavedCardUseCase(paymentRepo, savedCardRepo, provider)
//	payResponse, err := payUC.Execute(ctx, savedcard.PayWithSavedCardRequest{
//	    MemberID:   "member-uuid",
//	    CardID:     "card-uuid",
//	    Amount:     1999, // 19.99 in cents
//	    Currency:   "KZT",
//	    PaymentType: "subscription",
//	})
//	// payResponse contains: PaymentID, Status
//
// Saved card workflow:
//  1. Card saved during successful payment (via payment/save_card.go)
//  2. ListSavedCards: Display available cards to member
//  3. PayWithSavedCard: Charge selected card using stored token
//  4. DeleteSavedCard: Member removes unwanted card
//
// Card data stored:
//   - Card token (from provider, not actual card number)
//   - Masked card number (last 4 digits, e.g., "**** 1234")
//   - Card type (Visa, MasterCard, etc.)
//   - Expiry month/year
//   - Default flag (for one-click payments)
//
// Security:
//   - Only card tokens stored (never full card numbers)
//   - PCI DSS Level 1 compliance maintained
//   - Gateway handles all sensitive card data
//   - Tokens encrypted at rest
//   - Member ID validation prevents unauthorized access
//
// Architecture:
//   - Subdomain within payments bounded context
//   - Generic package name ("savedcard") within operations layer
//   - Import with alias: savedcardops "library-service/internal/payments/service/savedcard"
//   - DTOs colocated in http/savedcard/dto.go
package savedcard
