package domain

import (
	errors2 "library-service/internal/pkg/errors"
	"time"
)

// SubscriptionType represents different subscription tier types
type SubscriptionType string

const (
	SubscriptionBasic   SubscriptionType = "basic"
	SubscriptionPremium SubscriptionType = "premium"
	SubscriptionAnnual  SubscriptionType = "annual"
)

// Service encapsulates business logic for members and subscriptions
// This is a domain service in DDD terms
type Service struct {
	// Domain service are typically stateless
}

// NewService creates a new member domain service
func NewService() *Service {
	return &Service{}
}

// ValidateSubscriptionType checks if a subscription type is valid
func (s *Service) ValidateSubscriptionType(subType string) error {
	validTypes := map[string]bool{
		string(SubscriptionBasic):   true,
		string(SubscriptionPremium): true,
		string(SubscriptionAnnual):  true,
	}

	if !validTypes[subType] {
		return errors2.ErrInvalidInput.WithDetails("field", "subscription_type").
			WithDetails("valid_types", []string{"basic", "premium", "annual"})
	}

	return nil
}

// ValidateSubscriptionDuration checks if subscription duration is valid
func (s *Service) ValidateSubscriptionDuration(months int) error {
	const (
		minDuration = 1
		maxDuration = 24
	)

	if months < minDuration || months > maxDuration {
		return errors2.ErrInvalidInput.WithDetails("field", "duration_months").
			WithDetails("min", minDuration).
			WithDetails("max", maxDuration)
	}

	return nil
}

// CalculateSubscriptionPrice calculates the price for a subscription
// Business rule: Pricing based on type and duration with bulk discounts
func (s *Service) CalculateSubscriptionPrice(subType string, durationMonths int) (float64, error) {
	if err := s.ValidateSubscriptionType(subType); err != nil {
		return 0, err
	}

	if err := s.ValidateSubscriptionDuration(durationMonths); err != nil {
		return 0, err
	}

	// Base monthly prices
	monthlyPrices := map[string]float64{
		string(SubscriptionBasic):   9.99,
		string(SubscriptionPremium): 19.99,
		string(SubscriptionAnnual):  14.99, // Special annual rate
	}

	basePrice := monthlyPrices[subType]
	totalPrice := basePrice * float64(durationMonths)

	// Apply bulk discount for longer subscriptions
	// 10% discount for 6+ months, 20% for 12+ months
	if durationMonths >= 12 {
		totalPrice *= 0.80 // 20% discount
	} else if durationMonths >= 6 {
		totalPrice *= 0.90 // 10% discount
	}

	return totalPrice, nil
}

// CalculateExpirationDate calculates when a subscription will expire
func (s *Service) CalculateExpirationDate(startDate time.Time, durationMonths int) time.Time {
	return startDate.AddDate(0, durationMonths, 0)
}

// IsSubscriptionActive checks if a subscription is currently active
// Business rule: A subscription is active if current date is between start and expiration
func (s *Service) IsSubscriptionActive(subscribedAt, expiresAt time.Time) bool {
	now := time.Now()
	return now.After(subscribedAt) && now.Before(expiresAt)
}

// CanUpgradeSubscription checks if a member can upgrade their subscription
// Business rule: Can upgrade from basic to premium, but not downgrade
func (s *Service) CanUpgradeSubscription(currentType, targetType string) error {
	if err := s.ValidateSubscriptionType(currentType); err != nil {
		return err
	}
	if err := s.ValidateSubscriptionType(targetType); err != nil {
		return err
	}

	// Define subscription tier hierarchy
	tier := map[string]int{
		string(SubscriptionBasic):   1,
		string(SubscriptionPremium): 2,
		string(SubscriptionAnnual):  2, // Same tier as premium, different billing
	}

	if tier[targetType] < tier[currentType] {
		return errors2.ErrBusinessRule.WithDetails("reason", "Cannot downgrade subscription").
			WithDetails("current", currentType).
			WithDetails("target", targetType)
	}

	return nil
}

// ValidateMember validates member according to business rules
func (s *Service) Validate(member Member) error {
	if member.FullName == nil || *member.FullName == "" {
		return errors2.ErrInvalidMemberData.WithDetails("field", "full_name")
	}

	return nil
}

// CalculateGracePeriod calculates the grace period after subscription expiration
// Business rule: Premium members get 7 days grace, basic get 3 days
func (s *Service) CalculateGracePeriod(subType string) time.Duration {
	switch subType {
	case string(SubscriptionPremium), string(SubscriptionAnnual):
		return 7 * 24 * time.Hour // 7 days
	case string(SubscriptionBasic):
		return 3 * 24 * time.Hour // 3 days
	default:
		return 0
	}
}

// IsWithinGracePeriod checks if a member is within their grace period after expiration
func (s *Service) IsWithinGracePeriod(expiresAt time.Time, subType string) bool {
	now := time.Now()
	if now.Before(expiresAt) {
		return false // Still active, not in grace period
	}

	gracePeriod := s.CalculateGracePeriod(subType)
	graceEndDate := expiresAt.Add(gracePeriod)

	return now.Before(graceEndDate)
}
