// Package paymentops implements use cases for payment processing operations.
//
// This package orchestrates payment workflows for subscription fees, late fees,
// and other library charges. It integrates with external payment gateways
// (edomain.kz) and manages payment lifecycle from initiation to completion.
//
// Use cases implemented:
//   - InitiatePaymentUseCase: Starts payment process with gateway
//   - ProcessPaymentCallbackUseCase: Handles payment gateway callbacks
//   - GetPaymentStatusUseCase: Retrieves current payment status
//   - RefundPaymentUseCase: Processes payment refunds
//   - GenerateReceiptUseCase: Creates payment receipts for completed transactions
//   - ListMemberPaymentsUseCase: Returns payment history for a member
//
// Dependencies:
//   - domain.Repository: For payment record persistence
//   - domain.Service: For payment validation and receipt generation
//   - domain.Gateway: External payment gateway adapter (edomain.kz)
//   - domain.Repository: For updating member subscription status
//
// Example usage:
//
//	initiateUC := paymentops.NewInitiatePaymentUseCase(paymentRepo, gateway, memberRepo)
//	response, err := initiateUC.Execute(ctx, paymentops.InitiatePaymentRequest{
//	    MemberID:    "member-uuid",
//	    Amount:      999, // 9.99 in cents
//	    Type:        "subscription",
//	    Description: "Premium Subscription - Monthly",
//	})
//	// response contains: PaymentID, GatewayURL (redirect user here)
//
// Payment flow:
//  1. InitiatePayment: Create payment record, get gateway URL
//  2. User redirected to payment gateway
//  3. ProcessPaymentCallback: Verify signature, update payment status
//  4. GenerateReceipt: Create receipt for successful payments
//
// Payment types:
//   - Subscription: Monthly/annual membership fees
//   - Late fee: Overdue book return penalties
//   - Lost book: Book replacement charges
//
// Security:
//   - Payment gateway signatures verified (SHA-256 HMAC)
//   - Idempotency keys prevent duplicate charges
//   - All payment data encrypted at rest
//   - PCI DSS compliance maintained (no card data stored)
//
// Architecture:
//   - Package name uses "ops" suffix to avoid conflict with domain payment package
//   - Gateway adapter implements domain.Gateway interface
//   - Callback verification critical for security
//   - Failed payments tracked for retry logic
package payment
