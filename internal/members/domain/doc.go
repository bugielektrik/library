// Package member provides domain entities and business logic for member and subscription management.
//
// This package implements member-related domain logic including:
//   - Member entity with subscription tracking
//   - Subscription type validation (basic, premium, annual)
//   - Subscription pricing and discount calculation
//   - Expiration date calculation
//   - Grace period management
//   - Subscription upgrade/downgrade rules
//
// The member entity represents library members with their subscription status,
// book borrowing history, and access privileges.
//
// Example usage:
//
//	service := member.NewService()
//
//	// Calculate subscription price with discounts
//	price, err := service.CalculateSubscriptionPrice("premium", 12)
//	// Returns: 19.99 * 12 * 0.80 (20% discount for 12 months)
//
//	// Validate subscription upgrade
//	err = service.CanUpgradeSubscription("basic", "premium")
//
//	// Check grace period
//	inGrace := service.IsWithinGracePeriod(expirationDate, "premium")
//
// Subscription Types:
//   - basic: $9.99/month - 3-day grace period
//   - premium: $19.99/month - 7-day grace period
//   - annual: $14.99/month - 7-day grace period
//
// Pricing Discounts:
//   - 6+ months: 10% discount
//   - 12+ months: 20% discount
//
// Domain Rules:
//   - Member must have a valid full name (non-empty)
//   - Subscription duration: 1-24 months
//   - Can upgrade subscription tier (basic → premium → annual)
//   - Cannot downgrade subscription tier
//   - Grace period applies after expiration based on subscription type
package domain
