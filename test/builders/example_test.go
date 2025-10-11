package builders_test

import (
	"fmt"

	"library-service/internal/domain/payment"
	"library-service/test/builders"
)

// Example demonstrates basic builder usage
func Example() {
	// Create a payment with defaults
	p := builders.NewPayment().Build()

	fmt.Println("Status:", p.Status)
	fmt.Println("Amount:", p.Amount)

	// Output:
	// Status: pending
	// Amount: 10000
}

// ExamplePaymentBuilder demonstrates creating a payment with defaults
func ExamplePaymentBuilder() {
	payment := builders.NewPayment().Build()

	fmt.Println("Default status:", payment.Status)
	fmt.Println("Default currency:", payment.Currency)
	fmt.Println("Default amount:", payment.Amount)

	// Output:
	// Default status: pending
	// Default currency: KZT
	// Default amount: 10000
}

// ExamplePaymentBuilder_withCustomValues shows customizing payment fields
func ExamplePaymentBuilder_withCustomValues() {
	payment := builders.NewPayment().
		WithID("pay-123").
		WithAmount(50000).
		WithMemberID("member-456").
		Build()

	fmt.Println("ID:", payment.ID)
	fmt.Println("Amount:", payment.Amount)
	fmt.Println("Member ID:", payment.MemberID)

	// Output:
	// ID: pay-123
	// Amount: 50000
	// Member ID: member-456
}

// ExamplePaymentBuilder_WithCompletedStatus demonstrates completed payment
func ExamplePaymentBuilder_WithCompletedStatus() {
	payment := builders.NewPayment().
		WithID("pay-completed").
		WithCompletedStatus().
		Build()

	fmt.Println("Status:", payment.Status)
	fmt.Println("Has completion time:", payment.CompletedAt != nil)

	// Output:
	// Status: completed
	// Has completion time: true
}

// ExamplePaymentBuilder_WithFailedStatus demonstrates failed payment
func ExamplePaymentBuilder_WithFailedStatus() {
	payment := builders.NewPayment().
		WithID("pay-failed").
		WithFailedStatus().
		Build()

	fmt.Println("Status:", payment.Status)

	// Output:
	// Status: failed
}

// ExamplePaymentBuilder_chainedMethods demonstrates fluent interface
func ExamplePaymentBuilder_chainedMethods() {
	// Build complex payment in one chain
	payment := builders.NewPayment().
		WithID("pay-789").
		WithAmount(150000).
		WithCurrency(payment.CurrencyKZT).
		WithPaymentType(payment.PaymentTypeSubscription).
		WithPaymentMethod(payment.PaymentMethodCard).
		WithCardMask("****1234").
		WithCompletedStatus().
		Build()

	fmt.Println("Payment type:", payment.PaymentType)
	fmt.Println("Card mask:", *payment.CardMask)
	fmt.Println("Status:", payment.Status)

	// Output:
	// Payment type: subscription
	// Card mask: ****1234
	// Status: completed
}

// ExampleBookBuilder demonstrates book builder usage
func ExampleBookBuilder() {
	book := builders.NewBook().
		WithName("Clean Code").
		WithISBN("978-0132350884").
		WithAuthors("Robert C. Martin").
		Build()

	fmt.Println("Name:", *book.Name)
	fmt.Println("ISBN:", *book.ISBN)
	fmt.Println("Authors:", len(book.Authors))

	// Output:
	// Name: Clean Code
	// ISBN: 978-0132350884
	// Authors: 1
}

// ExampleMemberBuilder demonstrates member builder usage
func ExampleMemberBuilder() {
	member := builders.NewMember().
		WithEmail("test@example.com").
		WithFullName("Test User").
		Build()

	fmt.Println("Email:", member.Email)
	fmt.Println("Full name:", *member.FullName)
	fmt.Println("Role:", member.Role)

	// Output:
	// Email: test@example.com
	// Full name: Test User
	// Role: user
}

// Example_testScenario demonstrates using builders in test scenarios
func Example_testScenario() {
	// Scenario: Test payment processing for a premium subscription

	// 1. Create a member
	member := builders.NewMember().
		WithEmail("premium@example.com").
		Build()

	// 2. Create a pending payment for subscription
	pendingPayment := builders.NewPayment().
		WithMemberID(member.ID).
		WithAmount(15000). // 150.00
		WithPaymentType(payment.PaymentTypeSubscription).
		Build()

	fmt.Println("Pending payment:", pendingPayment.Status)

	// 3. Simulate successful payment
	completedPayment := builders.NewPayment().
		WithID(pendingPayment.ID).
		WithMemberID(member.ID).
		WithAmount(15000).
		WithCompletedStatus().
		Build()

	fmt.Println("Completed payment:", completedPayment.Status)
	fmt.Println("Same amount:", pendingPayment.Amount == completedPayment.Amount)

	// Output:
	// Pending payment: pending
	// Completed payment: completed
	// Same amount: true
}

// Example_builderPatternBenefits demonstrates why builders are useful
func Example_builderPatternBenefits() {
	// Without builder (verbose, error-prone):
	// payment := payment.Payment{
	//     ID: "test-id",
	//     InvoiceID: "test-invoice",
	//     MemberID: "test-member",
	//     Amount: 10000,
	//     Currency: payment.CurrencyKZT,
	//     Status: payment.StatusPending,
	//     ... 10+ more fields
	// }

	// With builder (concise, only specify what you need):
	payment := builders.NewPayment().
		WithAmount(25000).
		Build()

	// All other fields have sensible defaults
	fmt.Println("Amount customized:", payment.Amount == 25000)
	fmt.Println("Has default ID:", payment.ID != "")
	fmt.Println("Has default currency:", payment.Currency != "")

	// Output:
	// Amount customized: true
	// Has default ID: true
	// Has default currency: true
}
