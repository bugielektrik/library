# ADR 010: Domain Service for Payment Status Management

**Status:** Accepted
**Date:** 2025-10-09
**Context:** Phase 3 Refactoring - Clean Architecture Compliance

## Context

Payment status management logic was incorrectly placed in the **use case layer** (`internal/usecase/paymentops/handle_callback.go`) rather than the **domain layer**. This violated Clean Architecture's dependency rule and separated business logic from domain entities.

### Problems Identified

**1. Business Logic in Use Case Layer**
```go
// handle_callback.go (USE CASE LAYER) ❌
func (uc *HandleCallbackUseCase) mapGatewayStatus(gatewayStatus string) payment.Status {
    switch gatewayStatus {
    case "success", "approved":
        return payment.StatusCompleted
    case "failed", "declined":
        return payment.StatusFailed
    // ... business rules in orchestration layer
    }
}
```

**Why this is wrong:**
- ❌ Business rules (status mapping) in orchestration layer
- ❌ Duplicated logic if another use case needs status mapping
- ❌ Domain entity (`payment.Payment`) can't validate its own state transitions
- ❌ Testing business rules requires mocking repositories

**2. Scattered Payment Update Logic**
```go
// handle_callback.go (USE CASE LAYER) ❌
// Manual field updates spread across use case
payment.Status = newStatus
payment.GatewayTransactionID = &req.TransactionID
payment.CardMask = req.CardMask
payment.ApprovalCode = req.ApprovalCode
// ... 10 more field assignments
payment.UpdatedAt = time.Now()
if newStatus == payment.StatusCompleted {
    now := time.Now()
    payment.CompletedAt = &now
}
```

**Why this is problematic:**
- ❌ No encapsulation of update rules
- ❌ Easy to forget fields (e.g., UpdatedAt, CompletedAt)
- ❌ Business rule (set CompletedAt when status=Completed) buried in use case
- ❌ Hard to test update logic in isolation

**3. Final Status Check Duplication**
```go
// handle_callback.go (USE CASE LAYER) ❌
if payment.Status == payment.StatusCompleted ||
   payment.Status == payment.StatusRefunded ||
   payment.Status == payment.StatusCancelled {
    // Idempotency: Don't process final states
}
```

**Why this is problematic:**
- ❌ Business rule (what states are final) in use case
- ❌ Duplicated if another use case needs this check
- ❌ No single source of truth for "final states"

## Decision

**Extract all payment status management business logic to the domain service layer.**

Created 3 new methods in `internal/domain/payment/service.go`:

### 1. MapGatewayStatus() - Status Translation

**Purpose:** Convert external payment gateway status strings to internal domain `Status` enum

```go
// service.go (DOMAIN LAYER) ✅
func (s *Service) MapGatewayStatus(gatewayStatus string) Status {
    switch gatewayStatus {
    case "success", "approved":
        return StatusCompleted
    case "failed", "declined":
        return StatusFailed
    case "cancelled":
        return StatusCancelled
    case "processing":
        return StatusProcessing
    default:
        return StatusFailed  // Defensive: unknown → failed
    }
}
```

**Benefits:**
- ✅ Business rule: "unknown statuses default to Failed" documented
- ✅ Single source of truth for gateway status mapping
- ✅ Testable without mocking database
- ✅ Reusable across all payment use cases

### 2. IsFinalStatus() - Terminal State Detection

**Purpose:** Determine if a payment is in a state that cannot transition further

```go
// service.go (DOMAIN LAYER) ✅
func (s *Service) IsFinalStatus(status Status) bool {
    return status == StatusCompleted ||
           status == StatusRefunded ||
           status == StatusCancelled
}
```

**Benefits:**
- ✅ Single source of truth for final states
- ✅ Idempotency logic moved to domain
- ✅ Easy to test business rule

### 3. UpdateStatusFromCallback() - Encapsulated Update Logic

**Purpose:** Update payment entity from callback data following business rules

```go
// service.go (DOMAIN LAYER) ✅
type CallbackData struct {
    TransactionID   string
    CardMask        *string
    ApprovalCode    *string
    ErrorCode       *string
    ErrorMessage    *string
    GatewayResponse *string
    NewStatus       Status
}

func (s *Service) UpdateStatusFromCallback(payment *Payment, data CallbackData) {
    // Update status
    payment.Status = data.NewStatus

    // Update provider fields
    payment.GatewayTransactionID = &data.TransactionID
    payment.CardMask = data.CardMask
    payment.ApprovalCode = data.ApprovalCode
    payment.ErrorCode = data.ErrorCode
    payment.ErrorMessage = data.ErrorMessage
    payment.GatewayResponse = data.GatewayResponse

    // Update timestamps
    payment.UpdatedAt = time.Now()

    // Business rule: Set completion timestamp for completed payments
    if data.NewStatus == StatusCompleted {
        now := time.Now()
        payment.CompletedAt = &now
    }
}
```

**Benefits:**
- ✅ All update rules in one place
- ✅ Impossible to forget UpdatedAt or CompletedAt
- ✅ Business rule: "CompletedAt set when status=Completed" explicit
- ✅ Testable in isolation

## Refactored Use Case

### Before (Business Logic in Use Case)
```go
// handle_callback.go (96 lines) ❌
func (uc *HandleCallbackUseCase) Execute(ctx context.Context, req PaymentCallbackRequest) {
    // ... validation code

    // ❌ Final status check in use case
    if payment.Status == payment.StatusCompleted ||
       payment.Status == payment.StatusRefunded ||
       payment.Status == payment.StatusCancelled {
        // Skip processing
    }

    // ❌ Status mapping in use case
    newStatus := uc.mapGatewayStatus(req.Status)

    // ❌ Manual field updates in use case
    payment.Status = newStatus
    payment.GatewayTransactionID = &req.TransactionID
    payment.CardMask = req.CardMask
    payment.ApprovalCode = req.ApprovalCode
    // ... 10 more assignments
    payment.UpdatedAt = time.Now()
    if newStatus == payment.StatusCompleted {
        now := time.Now()
        payment.CompletedAt = &now
    }

    // ... save to repository
}

// ❌ Business logic method in use case
func (uc *HandleCallbackUseCase) mapGatewayStatus(status string) payment.Status {
    // ... business rules here
}
```

### After (Delegating to Domain Service)
```go
// handle_callback.go (84 lines) ✅
func (uc *HandleCallbackUseCase) Execute(ctx context.Context, req PaymentCallbackRequest) {
    // ... validation code

    // ✅ Delegate to domain service
    if uc.paymentService.IsFinalStatus(paymentEntity.Status) {
        // Skip processing
    }

    // ✅ Delegate to domain service
    newStatus := uc.paymentService.MapGatewayStatus(req.Status)

    // ✅ Validate transition (already in domain service)
    if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, newStatus); err != nil {
        return HandleCallbackResponse{}, err
    }

    // ✅ Delegate to domain service
    uc.paymentService.UpdateStatusFromCallback(&paymentEntity, payment.CallbackData{
        TransactionID:   req.TransactionID,
        CardMask:        req.CardMask,
        ApprovalCode:    req.ApprovalCode,
        ErrorCode:       req.ErrorCode,
        ErrorMessage:    req.ErrorMessage,
        GatewayResponse: gatewayResponseStr,
        NewStatus:       newStatus,
    })

    // ... save to repository
}
```

**Result:** Use case becomes pure orchestration, business logic in domain

## Architecture Alignment

### Clean Architecture Dependency Rule

```
┌─────────────────────────────────────────────┐
│ Use Case Layer (Orchestration)             │
│ ✅ Calls domain service                     │
│ ✅ Does NOT contain business logic          │
└────────────┬────────────────────────────────┘
             │ depends on ↓
┌────────────▼────────────────────────────────┐
│ Domain Layer (Business Logic)               │
│ ✅ Contains all payment status rules        │
│ ✅ MapGatewayStatus()                       │
│ ✅ IsFinalStatus()                          │
│ ✅ UpdateStatusFromCallback()               │
│ ✅ ValidateStatusTransition() (already had) │
│ ✅ NO external dependencies                 │
└─────────────────────────────────────────────┘
```

### Before vs After

| Concern | Before | After |
|---------|--------|-------|
| Status mapping | ❌ Use case layer | ✅ Domain service |
| Final status check | ❌ Use case layer | ✅ Domain service |
| Update logic | ❌ Use case layer | ✅ Domain service |
| Timestamp management | ❌ Use case layer | ✅ Domain service |
| Testing business rules | ❌ Needs DB mocks | ✅ Pure functions |

## Consequences

### Positive

✅ **Clean Architecture compliance:** Business logic properly in domain layer
✅ **Single source of truth:** All status rules in one place
✅ **Testability:** Domain service tests don't need database
✅ **Reusability:** Other use cases can use same methods
✅ **Maintainability:** Business rule changes only touch domain layer
✅ **Encapsulation:** Payment entity manages its own state transitions

### Neutral

⚠️ **More delegation:** Use case makes more domain service calls (acceptable tradeoff)
⚠️ **CallbackData struct:** New DTO for update operations (improves API clarity)

### Negative

❌ **None identified:** Pure architectural improvement

## Testing Strategy

### Domain Service Tests (No Database Required)

```go
func TestMapGatewayStatus(t *testing.T) {
    s := payment.NewService()

    tests := []struct {
        gatewayStatus string
        expected      payment.Status
    }{
        {"success", payment.StatusCompleted},
        {"approved", payment.StatusCompleted},
        {"failed", payment.StatusFailed},
        {"declined", payment.StatusFailed},
        {"cancelled", payment.StatusCancelled},
        {"processing", payment.StatusProcessing},
        {"unknown-status", payment.StatusFailed}, // Defensive default
    }

    for _, tt := range tests {
        got := s.MapGatewayStatus(tt.gatewayStatus)
        if got != tt.expected {
            t.Errorf("MapGatewayStatus(%q) = %v, want %v", tt.gatewayStatus, got, tt.expected)
        }
    }
}

func TestIsFinalStatus(t *testing.T) {
    s := payment.NewService()

    finalStatuses := []payment.Status{
        payment.StatusCompleted,
        payment.StatusRefunded,
        payment.StatusCancelled,
    }

    for _, status := range finalStatuses {
        if !s.IsFinalStatus(status) {
            t.Errorf("Expected %v to be final status", status)
        }
    }

    nonFinalStatuses := []payment.Status{
        payment.StatusPending,
        payment.StatusProcessing,
        payment.StatusFailed,
    }

    for _, status := range nonFinalStatuses {
        if s.IsFinalStatus(status) {
            t.Errorf("Expected %v to NOT be final status", status)
        }
    }
}

func TestUpdateStatusFromCallback(t *testing.T) {
    s := payment.NewService()
    p := payment.Payment{Status: payment.StatusProcessing}

    s.UpdateStatusFromCallback(&p, payment.CallbackData{
        TransactionID: "txn-123",
        NewStatus:     payment.StatusCompleted,
    })

    // Verify business rules
    assert.Equal(t, payment.StatusCompleted, p.Status)
    assert.Equal(t, "txn-123", *p.GatewayTransactionID)
    assert.NotNil(t, p.CompletedAt) // ✅ Business rule: CompletedAt set
    assert.NotZero(t, p.UpdatedAt)  // ✅ Business rule: UpdatedAt set
}
```

**Benefits:**
- ✅ Fast tests (no database, no network)
- ✅ Test business rules directly
- ✅ 100% coverage achievable

### Use Case Tests (With Mocks)

```go
func TestHandleCallback_UsesDomainService(t *testing.T) {
    // Verify use case delegates to domain service
    // Mock domain service to verify calls
    // Test orchestration, not business logic
}
```

## Migration Impact

**For existing code:** ✅ **ZERO BREAKING CHANGES**

Public API of `HandleCallbackUseCase` remains unchanged:
```go
// Before and After - same signature
func (uc *HandleCallbackUseCase) Execute(ctx context.Context, req PaymentCallbackRequest) (HandleCallbackResponse, error)
```

## Related ADRs

- **ADR 002: Domain Services** - Justifies domain service pattern
- **ADR 001: Clean Architecture** - Dependency rule compliance
- **ADR 009: Payment Gateway Modularization** - Gateway provides callback data

## Examples

### Example 1: Adding New Gateway Status

**Before:** Change use case layer ❌
```go
// handle_callback.go
func (uc *HandleCallbackUseCase) mapGatewayStatus(status string) payment.Status {
    switch status {
    case "success", "approved":
        return payment.StatusCompleted
    case "pending_review":  // ❌ New status added in use case
        return payment.StatusProcessing
    }
}
```

**After:** Change domain service ✅
```go
// service.go (DOMAIN LAYER)
func (s *Service) MapGatewayStatus(gatewayStatus string) Status {
    switch gatewayStatus {
    case "success", "approved":
        return StatusCompleted
    case "pending_review":  // ✅ New status added in domain
        return StatusProcessing
    }
}
```

**Result:** Business logic change properly in domain layer

### Example 2: Reusing Status Logic

**Before:** Duplicate logic ❌
```go
// Another use case needs final status check
func (uc *RefundPaymentUseCase) Execute(...) {
    // ❌ Duplicate final status check
    if payment.Status == payment.StatusCompleted ||
       payment.Status == payment.StatusRefunded ||
       payment.Status == payment.StatusCancelled {
        // ...
    }
}
```

**After:** Reuse domain service ✅
```go
// Another use case reuses domain service
func (uc *RefundPaymentUseCase) Execute(...) {
    // ✅ Reuse domain service method
    if uc.paymentService.IsFinalStatus(payment.Status) {
        // ...
    }
}
```

**Result:** DRY principle, single source of truth

## Lessons Learned

1. **Business logic belongs in domain:** Status mapping, state validation
2. **Use cases orchestrate:** Call domain services, don't implement logic
3. **Encapsulation improves maintainability:** UpdateStatusFromCallback prevents field omissions
4. **Domain services are highly testable:** No infrastructure dependencies
5. **Refactoring improves architecture:** Moving logic to correct layer clarifies responsibilities

## Future Work

**Additional domain service methods to consider:**

1. **CanRefund(payment, refundPolicy)** - Already exists, good example
2. **CalculateRefundAmount(payment, partialAmount)** - Refund validation
3. **IsRetryable(payment)** - Can failed payment be retried?
4. **GetNextAllowedStatuses(currentStatus)** - Valid next states

## Conclusion

Extracting payment status management to the domain service layer restored Clean Architecture compliance, improved testability, and created a single source of truth for payment business rules. The use case layer now focuses on orchestration (calling services, repositories) rather than implementing business logic.

**Impact:**
- ✅ 3 new domain service methods
- ✅ 12 lines removed from use case (business logic → domain)
- ✅ All tests passing (11 use case tests + domain service tests)
- ✅ Zero breaking changes to public API
- ✅ Clean Architecture dependency rule satisfied
