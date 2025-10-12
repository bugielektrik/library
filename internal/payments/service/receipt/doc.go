// Package receipt implements use cases for payment receipt generation and management.
//
// This package orchestrates workflows for creating, retrieving, and listing payment
// receipts. Receipts serve as official records of completed transactions and are
// required for accounting, tax reporting, and member transaction history.
//
// Use cases implemented:
//   - GenerateReceiptUseCase: Creates receipt for successful payment
//   - GetReceiptUseCase: Retrieves specific receipt by ID
//   - ListReceiptsUseCase: Returns all receipts for a member
//
// Dependencies:
//   - domain.ReceiptRepository: For receipt persistence
//   - domain.Repository: For accessing payment details
//   - domain.Repository (member): For member information
//
// Example usage:
//
//	generateUC := receipt.NewGenerateReceiptUseCase(paymentRepo, receiptRepo, memberRepo)
//	response, err := generateUC.Execute(ctx, receipt.GenerateReceiptRequest{
//	    PaymentID: "payment-uuid",
//	    MemberID:  "member-uuid",
//	    Notes:     "Annual subscription renewal",
//	})
//	// response contains: ReceiptID, ReceiptNumber (e.g., "RCP-2025-000123")
//
//	getUC := receipt.NewGetReceiptUseCase(receiptRepo)
//	receipt, err := getUC.Execute(ctx, receipt.GetReceiptRequest{
//	    ReceiptID: "receipt-uuid",
//	    MemberID:  "member-uuid",
//	})
//	// receipt contains: Full receipt with itemized charges, totals, timestamps
//
// Receipt generation workflow:
//  1. Payment completes successfully
//  2. GenerateReceipt: Create official receipt record
//  3. Assign unique receipt number (RCP-YYYY-NNNNNN format)
//  4. Store receipt with payment details, member info, timestamps
//  5. Member can retrieve/download receipt anytime
//
// Receipt data includes:
//   - Unique receipt number (sequential, year-prefixed)
//   - Payment ID and transaction reference
//   - Member information (name, email)
//   - Itemized charges with descriptions
//   - Amount, currency, payment method
//   - Tax information (if applicable)
//   - Timestamps (issued date, payment date)
//   - Status (issued, voided)
//
// Receipt numbering:
//   - Format: RCP-{YEAR}-{SEQUENCE}
//   - Example: RCP-2025-000123
//   - Sequential per year for accounting
//   - Unique constraint enforced at database level
//
// Use cases:
//   - Transaction history for members
//   - Accounting and financial reporting
//   - Tax documentation
//   - Dispute resolution
//   - Audit trails
//
// Architecture:
//   - Subdomain within payments bounded context
//   - Generic package name ("receipt") within operations layer
//   - Import with alias: receiptops "library-service/internal/payments/service/receipt"
//   - DTOs colocated in http/receipt/dto.go
//   - Receipts immutable once generated (audit requirement)
package receipt
