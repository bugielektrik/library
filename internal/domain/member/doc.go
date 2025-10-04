// Package member provides the core business logic and entities for library membership and subscriptions.
//
// This package implements the domain layer of Clean Architecture, containing:
//   - Member entity with subscription management
//   - Domain service for subscription validation and pricing
//   - Repository interface for data persistence
//   - Business rules for subscription lifecycle
//
// Subscription Types:
//   - basic: $9.99/month, 3-day grace period, can borrow 5 books
//   - premium: $19.99/month, 7-day grace period, unlimited books
//
// Business Rules:
//   - Subscription type must be 'basic' or 'premium'
//   - Duration must be 1-24 months
//   - Bulk discounts: 6+ months (10% off), 12+ months (20% off)
//   - Cannot downgrade subscription (premium → basic blocked)
//   - Grace period applies after expiration
//
// Example Usage:
//
//	// Create domain service
//	service := member.NewService()
//
//	// Calculate subscription price with bulk discount
//	price, err := service.CalculateSubscriptionPrice("premium", 12)
//	// Returns: $191.90 (20% discount on $239.88)
//
//	// Check if subscription is within grace period
//	expiresAt := time.Now().Add(-2 * 24 * time.Hour)
//	if service.IsWithinGracePeriod(expiresAt, "premium") {
//	    // Still within 7-day grace period
//	}
//
//	// Validate subscription upgrade
//	err := service.CanUpgradeSubscription("basic", "premium")
//	// Allowed
//
//	err = service.CanUpgradeSubscription("premium", "basic")
//	// Error: downgrade not allowed
//
// The member domain is independent of external frameworks and can be tested
// in isolation without database or HTTP dependencies.
package member
