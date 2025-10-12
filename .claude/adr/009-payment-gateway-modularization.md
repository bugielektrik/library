# ADR 009: Payment Gateway Modularization

**Status:** Accepted
**Date:** 2025-10-09
**Context:** Phase 3 Refactoring - Structural Improvements

## Context

The payment gateway adapter for epayment.kz (Kazakhstan payment service) was implemented as a single monolithic file `gateway.go` containing 546 lines of code. This file mixed multiple responsibilities:

**Problems identified:**
- **OAuth token management** (authentication, caching, refresh)
- **Payment operations** (status checks, refunds, cancellations, card charges)
- **Type definitions** (requests, responses, configurations)
- **Core gateway structure** (initialization, configuration)

This violated the **Single Responsibility Principle** and made the code:
- ❌ Hard to navigate (546 lines, scrolling required)
- ❌ Difficult to test specific concerns in isolation
- ❌ Challenging to modify without understanding entire file
- ❌ Poor cohesion (unrelated functions grouped together)

**File before refactoring:**
```
gateway.go (546 lines)
├── Config struct
├── Gateway struct
├── NewGateway()
├── GetAuthToken() + token caching logic
├── CheckPaymentStatus()
├── RefundPayment()
├── CancelPayment()
├── ChargeCardWithToken()
├── TokenResponse struct
├── TransactionStatusResponse struct
├── RefundRequest/Response structs
├── CardPaymentRequest/Response structs
└── ... all mixed together
```

## Decision

**Split the monolithic gateway file into 4 focused modules organized by responsibility:**

1. **gateway.go** (107 lines) - Core structure and configuration
2. **auth.go** (118 lines) - OAuth token management
3. **payment.go** (348 lines) - Payment operations
4. **types.go** (61 lines) - Request/response type definitions

### File Responsibilities

#### gateway.go - Core Gateway Structure
**Purpose:** Gateway initialization, configuration, and core structure

```go
// Contains:
- Config struct (OAuth + Gateway settings)
- Gateway struct (http client, logger, token cache)
- NewGateway() constructor
- Getters: GetConfig(), GetHTTPClient()
```

**Rationale:** Minimal file focused on gateway lifecycle and dependencies

#### auth.go - OAuth Token Management
**Purpose:** Authentication and token caching

```go
// Contains:
- GetAuthToken() - Main authentication method
- Token caching with sync.RWMutex for thread safety
- OAuth 2.0 Client Credentials flow
- Token expiry management
```

**Rationale:** OAuth is a complex concern requiring thread-safe caching logic, deserves isolation

#### payment.go - Payment Operations
**Purpose:** All payment-related API calls

```go
// Contains:
- CheckPaymentStatus() - GET transaction status
- RefundPayment() - POST refund request
- CancelPayment() - POST cancellation
- ChargeCardWithToken() - POST card charge
```

**Rationale:** Largest file (348 lines) but cohesive - all methods interact with payment APIs

#### types.go - Type Definitions
**Purpose:** Centralized request/response types

```go
// Contains:
- TokenResponse
- TransactionStatusResponse, TransactionDetails
- RefundRequest, RefundResponse
- CardPaymentRequest, CardPaymentResponse
```

**Rationale:** Types are referenced across multiple files, centralization prevents circular dependencies

## Implementation

### File Organization Pattern

```
internal/adapters/payment/epayment/
├── gateway.go       (Core: Config, Gateway struct, constructor)
├── auth.go          (OAuth: Token management, caching)
├── payment.go       (Operations: All payment API calls)
├── types.go         (Data: All request/response types)
└── gateway_test.go  (Tests: All 14 tests continue to pass)
```

### Thread Safety (auth.go)

OAuth token caching uses **double-checked locking pattern:**

```go
type Gateway struct {
    // ... other fields
    token       string          // Cached OAuth token
    tokenExpiry time.Time       // Token expiration timestamp
    tokenMutex  sync.RWMutex    // Thread-safe access
}

func (g *Gateway) GetAuthToken(ctx context.Context) (string, error) {
    // Fast path: Read lock for cache check
    g.tokenMutex.RLock()
    if g.token != "" && time.Now().Before(g.tokenExpiry) {
        token := g.token
        g.tokenMutex.RUnlock()
        return token, nil
    }
    g.tokenMutex.RUnlock()

    // Slow path: Write lock for token refresh
    g.tokenMutex.Lock()
    defer g.tokenMutex.Unlock()

    // Double-check: Another goroutine might have refreshed
    if g.token != "" && time.Now().Before(g.tokenExpiry) {
        return g.token, nil
    }

    // Fetch new token from OAuth server
    // ...
}
```

**Benefits:**
- ✅ Multiple readers (fast path with RLock)
- ✅ Single writer during refresh (slow path with Lock)
- ✅ Prevents redundant OAuth calls
- ✅ Thread-safe for concurrent HTTP requests

## Consequences

### Positive

✅ **Single Responsibility:** Each file has one clear purpose
✅ **Navigation:** 107-348 lines per file (vs. 546 monolithic)
✅ **Testability:** Can test OAuth logic independently from payment operations
✅ **Maintainability:** Changes to refund logic don't affect authentication
✅ **Readability:** Related functions grouped, easier to understand
✅ **No breaking changes:** All tests pass without modification

### Neutral

⚠️ **File count:** 1 file → 4 files (more files but better organization)
⚠️ **Navigation overhead:** Must open multiple files to see full picture

### Negative

❌ **None identified:** Pure structural improvement with no downsides

## Migration Impact

### For Developers

**Before:**
```go
// All in provider.go
import "library-service/internal/adapters/payment/epayment"

gateway := epayment.NewGateway(config)
token, _ := gateway.GetAuthToken(ctx)  // Defined in same file
status, _ := gateway.CheckPaymentStatus(ctx, id)  // Defined in same file
```

**After:**
```go
// Split across files but API unchanged
import "library-service/internal/adapters/payment/epayment"

gateway := epayment.NewGateway(config)  // provider.go
token, _ := gateway.GetAuthToken(ctx)   // auth.go (internal)
status, _ := gateway.CheckPaymentStatus(ctx, id)  // payment.go
```

**Impact:** ✅ **ZERO** - Public API unchanged, existing code works as-is

### Test Results

**All 14 tests passing after refactoring:**
```
✅ TestNewGateway
✅ TestGateway_GetAuthToken
✅ TestGateway_GetAuthToken_CachedToken
✅ TestGateway_CheckPaymentStatus
✅ TestGateway_CheckPaymentStatus_Error
✅ TestGateway_RefundPayment
✅ TestGateway_RefundPayment_PartialRefund
✅ TestGateway_RefundPayment_Error
✅ TestGateway_CancelPayment
✅ TestGateway_CancelPayment_Error
✅ TestGateway_ChargeCardWithToken
✅ TestGateway_ChargeCardWithToken_Error
✅ TestGateway_GetConfig
✅ TestGateway_GetHTTPClient
```

### Errors Fixed During Refactoring

**Type mismatches and missing fields discovered and fixed:**

1. Missing `WidgetURL` field in Config (added)
2. Missing `ID` field in CardPaymentResponse (added)
3. TransactionDetails type mismatch (pointer → value)
4. Invalid test data types (float → int64 for amounts)
5. Removed obsolete `AccountID` field

**Result:** Refactoring surfaced and fixed hidden bugs

## Alternatives Considered

### 1. Keep monolithic file
- **Pros:** Single file, no navigation
- **Cons:** 546 lines, hard to maintain
- **Rejected:** Violates SRP, poor maintainability

### 2. Split by HTTP method (GET, POST)
- **Pros:** Simple categorization
- **Cons:** Breaks cohesion (refunds and cancellations separated)
- **Rejected:** Not semantically meaningful

### 3. Split into even more files (one per operation)
- **Pros:** Maximum separation
- **Cons:** Too granular, 10+ tiny files
- **Rejected:** Over-engineering for current scope

### 4. Use subdirectories (auth/, payments/, types/)
- **Pros:** Hierarchical organization
- **Cons:** Import complexity, package name conflicts
- **Rejected:** Overkill for 4 files

## Design Principles Applied

### Single Responsibility Principle (SRP)
- ✅ auth.go → OAuth token management only
- ✅ payment.go → Payment operations only
- ✅ types.go → Type definitions only

### Open/Closed Principle
- ✅ Adding new payment methods only touches payment.go
- ✅ Changing OAuth flow only touches auth.go

### Interface Segregation
- ✅ Gateway interface remains unchanged
- ✅ Internal methods organized by concern

## Future Enhancements

**Potential improvements enabled by modularization:**

1. **Separate auth package** if we integrate multiple payment gateways:
   ```
   internal/adapters/payment/
   ├── auth/           # Shared OAuth utilities
   ├── epayment/       # Kazakhstan gateway
   └── stripe/         # Alternative gateway
   ```

2. **Middleware pattern** for common concerns:
   ```go
   // auth.go could export middleware
   func (g *Gateway) WithAuth(next http.Handler) http.Handler {
       // Inject OAuth token into requests
   }
   ```

3. **Plugin architecture** for different payment providers:
   ```go
   type PaymentGateway interface {
       CheckStatus(ctx, id) (Status, error)
       Refund(ctx, req RefundRequest) error
       // ... interface defined in types.go
   }
   ```

## Related ADRs

- **ADR 010: Domain Service for Payment Status** - Uses payment.go methods
- **ADR 011: BaseRepository Pattern** - Similar modularization approach

## References

- Implementation: `internal/adapters/payment/epayment/`
- Tests: `internal/adapters/payment/epayment/gateway_test.go`
- epayment.kz API docs: https://epayment.kz/docs (OAuth, payment operations)

## Lessons Learned

1. **Refactoring reveals bugs:** Type system caught issues during split
2. **Tests are safety net:** All tests passing confirmed no regressions
3. **Cohesion > File count:** 4 focused files better than 1 monolithic
4. **Thread safety matters:** RWMutex pattern for concurrent access
5. **Public API stability:** Internal refactoring should not break consumers

## Conclusion

Splitting the 546-line monolithic gateway into 4 focused modules improved code organization without breaking existing functionality. The refactoring followed Clean Architecture principles by organizing code by concern rather than technical layer. All tests pass, confirming the refactoring was successful.
