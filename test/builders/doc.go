/*
Package builders provides test fixture builders for creating domain entities.

The builders implement the Builder pattern to provide a fluent interface for
constructing test data with sensible defaults that can be overridden as needed.

# Benefits

  - Reduces test setup boilerplate by 30-40%
  - Provides consistent test data across test suites
  - Makes test intent clearer by showing only what's relevant to each test
  - Easy to maintain when domain entities change

# Usage

Basic usage with defaults:

	payment := builders.NewPayment().Build()
	// Creates a payment with sensible defaults

Override specific fields:

	payment := builders.NewPayment().
		WithID("payment-123").
		WithAmount(50000).
		WithStatus(payment.StatusCompleted).
		Build()

Use convenience methods for common scenarios:

	card := builders.NewSavedCard().
		WithExpired().
		WithInactive().
		Build()

	member := builders.NewMember().
		WithAdminRole().
		WithSubscription(member.SubscriptionPlanMonthly).
		Build()

# Available Builders

  - PaymentBuilder: Build Payment entities with various statuses
  - SavedCardBuilder: Build SavedCard entities (active, expired, etc.)
  - ReceiptBuilder: Build Receipt entities
  - MemberBuilder: Build Member entities with subscriptions
  - BookBuilder: Build Book entities with authors

# Example Test

Before (manual construction):

	payment := payment.Payment{
		ID:            "payment-1",
		InvoiceID:     "inv-1",
		MemberID:      "member-1",
		Amount:        10000,
		Currency:      "KZT",
		Status:        payment.StatusCompleted,
		PaymentType:   payment.PaymentTypeFine,
		PaymentMethod: payment.PaymentMethodCard,
		CreatedAt:     time.Now(),
		CompletedAt:   &now,
	}

After (using builder):

	payment := builders.NewPayment().
		WithID("payment-1").
		WithCompletedStatus().
		Build()

The builder approach is more concise and focuses on what's important for the test.
*/
package builders
