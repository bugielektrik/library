package member_test

import (
	"fmt"
	"time"

	"library-service/internal/domain/member"
)

// Example demonstrates basic member service usage
func Example() {
	svc := member.NewService()

	// Calculate subscription price
	price, _ := svc.CalculateSubscriptionPrice("premium", 12)
	fmt.Printf("12-month premium: $%.2f\n", price)

	// Output:
	// 12-month premium: $191.90
}

// ExampleService_CalculateSubscriptionPrice demonstrates pricing calculations
func ExampleService_CalculateSubscriptionPrice() {
	svc := member.NewService()

	// Basic subscription - 1 month
	price1, _ := svc.CalculateSubscriptionPrice("basic", 1)
	fmt.Printf("Basic 1 month: $%.2f\n", price1)

	// Premium subscription - 6 months (10% discount)
	price6, _ := svc.CalculateSubscriptionPrice("premium", 6)
	fmt.Printf("Premium 6 months: $%.2f\n", price6)

	// Annual subscription - 12 months (20% discount)
	price12, _ := svc.CalculateSubscriptionPrice("annual", 12)
	fmt.Printf("Annual 12 months: $%.2f\n", price12)

	// Output:
	// Basic 1 month: $9.99
	// Premium 6 months: $107.95
	// Annual 12 months: $143.90
}

// ExampleService_CalculateSubscriptionPrice_bulkDiscount shows discount tiers
func ExampleService_CalculateSubscriptionPrice_bulkDiscount() {
	svc := member.NewService()

	// No discount for < 6 months
	price3, _ := svc.CalculateSubscriptionPrice("premium", 3)
	fmt.Printf("3 months (no discount): $%.2f\n", price3)

	// 10% discount for 6-11 months
	price6, _ := svc.CalculateSubscriptionPrice("premium", 6)
	fmt.Printf("6 months (10%% off): $%.2f\n", price6)

	// 20% discount for 12+ months
	price12, _ := svc.CalculateSubscriptionPrice("premium", 12)
	fmt.Printf("12 months (20%% off): $%.2f\n", price12)

	// Output:
	// 3 months (no discount): $59.97
	// 6 months (10% off): $107.95
	// 12 months (20% off): $191.90
}

// ExampleService_IsSubscriptionActive demonstrates subscription status checks
func ExampleService_IsSubscriptionActive() {
	svc := member.NewService()

	now := time.Now()

	// Active subscription (started yesterday, expires in 30 days)
	startDate := now.AddDate(0, 0, -1)
	expiryDate := now.AddDate(0, 1, 0)
	active := svc.IsSubscriptionActive(startDate, expiryDate)
	fmt.Println("Active subscription:", active)

	// Expired subscription (expired yesterday)
	startDate2 := now.AddDate(0, -1, 0)
	expiryDate2 := now.AddDate(0, 0, -1)
	active2 := svc.IsSubscriptionActive(startDate2, expiryDate2)
	fmt.Println("Expired subscription:", active2)

	// Future subscription (starts tomorrow)
	startDate3 := now.AddDate(0, 0, 1)
	expiryDate3 := now.AddDate(0, 1, 1)
	active3 := svc.IsSubscriptionActive(startDate3, expiryDate3)
	fmt.Println("Future subscription:", active3)

	// Output:
	// Active subscription: true
	// Expired subscription: false
	// Future subscription: false
}

// ExampleService_CalculateExpirationDate shows expiration date calculation
func ExampleService_CalculateExpirationDate() {
	svc := member.NewService()

	startDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// 1 month subscription
	expiry1 := svc.CalculateExpirationDate(startDate, 1)
	fmt.Println("1 month expiry:", expiry1.Format("2006-01-02"))

	// 12 month subscription
	expiry12 := svc.CalculateExpirationDate(startDate, 12)
	fmt.Println("12 month expiry:", expiry12.Format("2006-01-02"))

	// Output:
	// 1 month expiry: 2024-02-15
	// 12 month expiry: 2025-01-15
}

// ExampleService_CanUpgradeSubscription demonstrates upgrade validation
func ExampleService_CanUpgradeSubscription() {
	svc := member.NewService()

	// Can upgrade from basic to premium
	err := svc.CanUpgradeSubscription("basic", "premium")
	fmt.Println("Basic to Premium:", err == nil)

	// Can upgrade from basic to annual
	err = svc.CanUpgradeSubscription("basic", "annual")
	fmt.Println("Basic to Annual:", err == nil)

	// Cannot downgrade from premium to basic
	err = svc.CanUpgradeSubscription("premium", "basic")
	fmt.Println("Premium to Basic (downgrade):", err != nil)

	// Output:
	// Basic to Premium: true
	// Basic to Annual: true
	// Premium to Basic (downgrade): true
}

// ExampleService_CalculateGracePeriod shows grace period calculation
func ExampleService_CalculateGracePeriod() {
	svc := member.NewService()

	// Basic members get 3 days grace period
	grace := svc.CalculateGracePeriod("basic")
	fmt.Printf("Basic grace period: %d days\n", int(grace.Hours()/24))

	// Premium members get 7 days grace period
	grace = svc.CalculateGracePeriod("premium")
	fmt.Printf("Premium grace period: %d days\n", int(grace.Hours()/24))

	// Annual members get 7 days grace period
	grace = svc.CalculateGracePeriod("annual")
	fmt.Printf("Annual grace period: %d days\n", int(grace.Hours()/24))

	// Output:
	// Basic grace period: 3 days
	// Premium grace period: 7 days
	// Annual grace period: 7 days
}

// ExampleService_ValidateSubscriptionType demonstrates type validation
func ExampleService_ValidateSubscriptionType() {
	svc := member.NewService()

	// Valid subscription types
	err := svc.ValidateSubscriptionType("basic")
	fmt.Println("basic valid:", err == nil)

	err = svc.ValidateSubscriptionType("premium")
	fmt.Println("premium valid:", err == nil)

	// Invalid subscription type
	err = svc.ValidateSubscriptionType("platinum")
	fmt.Println("platinum invalid:", err != nil)

	// Output:
	// basic valid: true
	// premium valid: true
	// platinum invalid: true
}

// ExampleService_ValidateSubscriptionDuration demonstrates duration validation
func ExampleService_ValidateSubscriptionDuration() {
	svc := member.NewService()

	// Valid durations (1-24 months)
	err := svc.ValidateSubscriptionDuration(1)
	fmt.Println("1 month valid:", err == nil)

	err = svc.ValidateSubscriptionDuration(12)
	fmt.Println("12 months valid:", err == nil)

	// Invalid durations
	err = svc.ValidateSubscriptionDuration(0)
	fmt.Println("0 months invalid:", err != nil)

	err = svc.ValidateSubscriptionDuration(25)
	fmt.Println("25 months invalid:", err != nil)

	// Output:
	// 1 month valid: true
	// 12 months valid: true
	// 0 months invalid: true
	// 25 months invalid: true
}

// Example_subscriptionLifecycle demonstrates complete subscription lifecycle
func Example_subscriptionLifecycle() {
	svc := member.NewService()

	// 1. Calculate price for 6-month premium subscription
	price, _ := svc.CalculateSubscriptionPrice("premium", 6)
	fmt.Printf("Price: $%.2f\n", price)

	// 2. Calculate expiration date (using current time for active subscription)
	startDate := time.Now().AddDate(0, 0, -30) // Started 30 days ago
	expiryDate := svc.CalculateExpirationDate(startDate, 6)
	monthsUntilExpiry := int(time.Until(expiryDate).Hours() / 24 / 30)
	fmt.Printf("Expires in approximately: %d months\n", monthsUntilExpiry)

	// 3. Check if subscription is active
	active := svc.IsSubscriptionActive(startDate, expiryDate)
	fmt.Printf("Active: %t\n", active)

	// 4. Get grace period
	grace := svc.CalculateGracePeriod("premium")
	fmt.Printf("Grace period: %d days\n", int(grace.Hours()/24))

	// Output:
	// Price: $107.95
	// Expires in approximately: 5 months
	// Active: true
	// Grace period: 7 days
}
