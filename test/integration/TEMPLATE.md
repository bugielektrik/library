# Integration Test Template

This document provides templates and patterns for writing integration tests.

## Basic Template

```go
//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"
    "library-service/test/fixtures"
)

// TestBookIntegration tests book operations with real database
func TestBookIntegration(t *testing.T) {
    // Setup: Get database connection
    db, cleanup := setupTestDB(t)
    defer cleanup()

    // Setup: Create repositories and use cases
    bookRepo := postgres.NewBookRepository(db)
    bookService := book.NewService()
    createUC := bookops.NewCreateBookUseCase(bookRepo, nil, bookService)

    ctx := context.Background()

    t.Run("CreateBook", func(t *testing.T) {
        // Arrange
        req := bookops.CreateBookRequest{
            Name:    "Integration Test Book",
            Genre:   "Technology",
            ISBN:    "978-" + generateRandomISBN(),
            Authors: []string{uuid.New().String()},
        }

        // Act
        result, err := createUC.Execute(ctx, req)

        // Assert
        require.NoError(t, err)
        assert.NotEmpty(t, result.ID)
        assert.Equal(t, req.Name, result.Name)

        // Verify in database
        entity, err := bookRepo.Get(ctx, result.ID)
        require.NoError(t, err)
        assert.Equal(t, req.ISBN, *entity.ISBN)
    })
}

// Helper function for test
func generateRandomISBN() string {
    return fmt.Sprintf("%010d", time.Now().UnixNano()%10000000000)
}
```

## Full Workflow Template

```go
//go:build integration

package integration

import (
    "context"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/payment"
    "library-service/internal/usecase/paymentops"
    "library-service/test/mocks"
)

// TestPaymentWorkflow tests complete payment flow
func TestPaymentWorkflow(t *testing.T) {
    // Setup
    db, cleanup := setupTestDB(t)
    defer cleanup()

    paymentRepo := postgres.NewPaymentRepository(db)
    paymentService := payment.NewService()
    mockGateway := mocks.NewPaymentGateway()

    initiateUC := paymentops.NewInitiatePaymentUseCase(paymentRepo, paymentService, mockGateway)
    handleCallbackUC := paymentops.NewHandleCallbackUseCase(paymentRepo, paymentService)

    ctx := context.Background()
    memberID := uuid.New().String()

    t.Run("CompletePaymentFlow", func(t *testing.T) {
        // Step 1: Initiate payment
        initiateReq := paymentops.InitiatePaymentRequest{
            MemberID:    memberID,
            Amount:      5000,
            Currency:    "KZT",
            PaymentType: payment.PaymentTypeFine,
        }

        initiateResp, err := initiateUC.Execute(ctx, initiateReq)
        require.NoError(t, err)
        require.NotEmpty(t, initiateResp.PaymentID)

        // Step 2: Verify payment created with pending status
        paymentEntity, err := paymentRepo.GetByID(ctx, initiateResp.PaymentID)
        require.NoError(t, err)
        require.Equal(t, payment.StatusPending, paymentEntity.Status)

        // Step 3: Simulate successful callback
        callbackReq := paymentops.PaymentCallbackRequest{
            InvoiceID:     initiateResp.InvoiceID,
            TransactionID: "txn-" + uuid.New().String(),
            Amount:        5000,
            Currency:      "KZT",
            Status:        "success",
        }

        callbackResp, err := handleCallbackUC.Execute(ctx, callbackReq)
        require.NoError(t, err)
        require.Equal(t, payment.StatusCompleted, callbackResp.Status)

        // Step 4: Verify final state
        updatedPayment, err := paymentRepo.GetByID(ctx, initiateResp.PaymentID)
        require.NoError(t, err)
        require.Equal(t, payment.StatusCompleted, updatedPayment.Status)
        require.NotNil(t, updatedPayment.CompletedAt)
    })
}
```

## Parallel Test Template

```go
//go:build integration

package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/test/fixtures"
)

// TestParallelOperations tests operations that can run in parallel
func TestParallelOperations(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    bookRepo := postgres.NewBookRepository(db)
    ctx := context.Background()

    t.Run("CreateMultipleBooks", func(t *testing.T) {
        tests := []struct {
            name string
            isbn string
        }{
            {"Book1", "978-0001"},
            {"Book2", "978-0002"},
            {"Book3", "978-0003"},
        }

        for _, tt := range tests {
            tt := tt // Capture range variable
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel() // Run in parallel

                book := fixtures.ValidBook()
                book.ISBN = &tt.isbn

                id, err := bookRepo.Add(ctx, book)
                require.NoError(t, err)
                require.NotEmpty(t, id)
            })
        }
    })
}
```

## Error Scenario Template

```go
//go:build integration

package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/payment"
    "library-service/internal/usecase/paymentops"
    "library-service/internal/infrastructure/pkg/errors"
)

// TestPaymentErrorScenarios tests error handling
func TestPaymentErrorScenarios(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    paymentRepo := postgres.NewPaymentRepository(db)
    paymentService := payment.NewService()
    refundUC := paymentops.NewRefundPaymentUseCase(paymentRepo, paymentService, nil)

    ctx := context.Background()

    t.Run("RefundNonExistentPayment", func(t *testing.T) {
        req := paymentops.RefundPaymentRequest{
            PaymentID: "non-existent-id",
            MemberID:  "member-id",
            Reason:    "Test",
            IsAdmin:   true,
        }

        _, err := refundUC.Execute(ctx, req)
        require.Error(t, err)
        assert.ErrorIs(t, err, errors.ErrNotFound)
    })

    t.Run("RefundAlreadyRefundedPayment", func(t *testing.T) {
        // Create refunded payment
        refundedPayment := payment.Payment{
            ID:       uuid.New().String(),
            MemberID: "member-id",
            Amount:   1000,
            Status:   payment.StatusRefunded,
            // ... other fields
        }

        paymentID, err := paymentRepo.Create(ctx, refundedPayment)
        require.NoError(t, err)

        // Try to refund again
        req := paymentops.RefundPaymentRequest{
            PaymentID: paymentID,
            MemberID:  "member-id",
            Reason:    "Test",
            IsAdmin:   true,
        }

        _, err = refundUC.Execute(ctx, req)
        require.Error(t, err)
        assert.ErrorIs(t, err, errors.ErrInvalidPaymentStatus)
    })
}
```

## Time-Dependent Test Template

```go
//go:build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/payment"
    "library-service/internal/usecase/paymentops"
)

// TestPaymentExpiration tests time-dependent expiration logic
func TestPaymentExpiration(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    paymentRepo := postgres.NewPaymentRepository(db)
    paymentService := payment.NewService()
    expireUC := paymentops.NewExpirePaymentsUseCase(paymentRepo, paymentService)

    ctx := context.Background()

    t.Run("ExpirePendingPayments", func(t *testing.T) {
        // Create expired payment
        expiredPayment := payment.Payment{
            ID:        uuid.New().String(),
            MemberID:  uuid.New().String(),
            InvoiceID: "expired-" + uuid.New().String(),
            Amount:    2000,
            Status:    payment.StatusPending,
            CreatedAt: time.Now().Add(-2 * time.Hour),
            ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
            // ... other fields
        }

        paymentID, err := paymentRepo.Create(ctx, expiredPayment)
        require.NoError(t, err)

        // Run expiration
        req := paymentops.ExpirePaymentsRequest{
            BatchSize: 100,
        }

        resp, err := expireUC.Execute(ctx, req)
        require.NoError(t, err)
        require.Equal(t, 1, resp.ExpiredCount)

        // Verify status changed
        updated, err := paymentRepo.GetByID(ctx, paymentID)
        require.NoError(t, err)
        require.Equal(t, payment.StatusFailed, updated.Status)
    })
}
```

## Idempotency Test Template

```go
//go:build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/payment"
    "library-service/internal/usecase/paymentops"
)

// TestPaymentIdempotency tests idempotent operations
func TestPaymentIdempotency(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    paymentRepo := postgres.NewPaymentRepository(db)
    paymentService := payment.NewService()
    handleCallbackUC := paymentops.NewHandleCallbackUseCase(paymentRepo, paymentService)

    ctx := context.Background()

    t.Run("DuplicateCallbackIgnored", func(t *testing.T) {
        // Create completed payment
        completedPayment := payment.Payment{
            ID:        uuid.New().String(),
            MemberID:  uuid.New().String(),
            InvoiceID: "idempotency-" + uuid.New().String(),
            Amount:    3000,
            Status:    payment.StatusCompleted,
            CreatedAt: time.Now(),
            ExpiresAt: time.Now().Add(30 * time.Minute),
            // ... other fields
        }

        paymentID, err := paymentRepo.Create(ctx, completedPayment)
        require.NoError(t, err)

        // Send duplicate callback
        callbackReq := paymentops.PaymentCallbackRequest{
            InvoiceID:     completedPayment.InvoiceID,
            TransactionID: "txn-" + uuid.New().String(),
            Amount:        3000,
            Currency:      "KZT",
            Status:        "success",
        }

        resp, err := handleCallbackUC.Execute(ctx, callbackReq)
        require.NoError(t, err)
        assert.Equal(t, paymentID, resp.PaymentID)
        assert.Equal(t, payment.StatusCompleted, resp.Status)
        assert.False(t, resp.Processed) // Should indicate no processing
    })
}
```

## Best Practices

### ✅ Do

1. **Use build tags**: Always include `//go:build integration`
2. **Setup/cleanup**: Use `setupTestDB(t)` with `defer cleanup()`
3. **Unique data**: Use UUIDs or timestamps for unique test data
4. **Verify state**: Check database state after operations
5. **Test workflows**: Test complete end-to-end flows
6. **Test errors**: Include error scenarios and edge cases
7. **Use require**: Use `require` for critical checks, `assert` for non-critical
8. **Parallel tests**: Use `t.Parallel()` when tests don't conflict
9. **Mock externals**: Mock external services, use real DB
10. **Clean names**: Use descriptive test and subtest names

### ❌ Don't

1. **Don't skip cleanup**: Always defer cleanup functions
2. **Don't share state**: Each test should be independent
3. **Don't hardcode IDs**: Use UUIDs or generated IDs
4. **Don't test units**: Integration tests are for workflows
5. **Don't mock DB**: Use real database in integration tests
6. **Don't ignore errors**: Always check error returns
7. **Don't use time.Sleep**: Use actual state checks instead
8. **Don't commit test data**: Clean up after tests

## Running Integration Tests

```bash
# All integration tests
make test-integration

# Specific test
go test -tags integration -v -run TestPaymentWorkflow ./test/integration/

# With coverage
go test -tags integration -coverprofile=coverage.out ./test/integration/

# Verbose with race detection
go test -tags integration -v -race ./test/integration/
```

## Debugging Integration Tests

### Enable Verbose Output
```bash
go test -tags integration -v ./test/integration/
```

### Run Single Test
```bash
go test -tags integration -v -run TestPaymentWorkflow/CompletePaymentFlow ./test/integration/
```

### Check Database State
```sql
-- After test failure, check database
SELECT * FROM payments WHERE invoice_id = 'test-invoice-id';
SELECT * FROM receipts WHERE payment_id = 'test-payment-id';
```

### Enable SQL Logging
```go
// In setupTestDB
db, err := sqlx.Open("postgres", dsn)
if err != nil {
    t.Fatal(err)
}

// Enable query logging for debugging
db.SetMaxOpenConns(1) // Force serial execution for debugging
```

## Related Documentation

- [Test README](../README.md) - Test infrastructure overview
- [Testing Guide](../../.claude/testing.md) - Comprehensive testing strategies
- [Development Workflows](../../.claude/development-workflows.md) - Complete workflows
