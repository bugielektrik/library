# ADR 005: Payment Gateway Interface Abstraction

**Status:** Accepted

**Date:** 2025-10-09

**Context:**

Payment processing requires integration with external payment gateways (epayment.kz in our case). Direct coupling to a specific gateway creates problems:

```go
// BAD: Use case directly coupled to epayment.kz
import "library-service/internal/adapters/payment/epayment"

type InitiatePaymentUseCase struct {
    epaymentGateway *epayment.Gateway  // Tightly coupled!
}

func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, req Request) (Response, error) {
    // Directly using epayment-specific types and methods
    result, err := uc.epaymentGateway.InitiatePayment(...)
    // ...
}
```

**Problems:**
1. **Tight coupling:** Use cases depend on specific gateway implementation
2. **Hard to test:** Can't mock payment gateway easily
3. **Hard to switch:** Changing gateways requires modifying all use cases
4. **Vendor lock-in:** Business logic tied to epayment.kz specifics

## Decision

Define a **gateway-agnostic interface** in the use case layer:

```
Use Case Layer
    ↓ defines
PaymentGateway Interface
    ↑ implements
Adapter Layer
    ↓
epayment.Gateway (concrete implementation)
```

## Implementation

### Interface Definition

```go
// internal/usecase/container.go
type PaymentGateway interface {
    GetAuthToken(ctx context.Context) (string, error)
    GetTerminal() string
    GetBackLink() string
    GetPostLink() string
    GetWidgetURL() string
    CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error)
}
```

**Key characteristics:**
- **Minimal:** Only methods use cases actually need
- **Gateway-agnostic:** No epayment-specific types
- **Located in use case layer:** Use cases own the interface
- **Implementation in adapter layer:** epayment package implements it

### Use Case Usage

```go
// internal/usecase/paymentops/initiate_payment.go
type InitiatePaymentUseCase struct {
    paymentRepo    payment.Repository
    paymentService *payment.Service
    gateway        PaymentGateway  // Interface, not concrete type!
}

func (uc *InitiatePaymentUseCase) Execute(ctx context.Context, req Request) (Response, error) {
    // Use gateway methods without knowing implementation
    token, err := uc.gateway.GetAuthToken(ctx)
    if err != nil {
        return Response{}, fmt.Errorf("failed to get auth token: %w", err)
    }

    return Response{
        AuthToken: token,
        Terminal:  uc.gateway.GetTerminal(),
        BackLink:  uc.gateway.GetBackLink(),
        WidgetURL: uc.gateway.GetWidgetURL(),
    }, nil
}
```

### Implementation in Adapter

```go
// internal/adapters/payment/epayment/gateway.go
package epayment

type Gateway struct {
    config     *Config
    httpClient *http.Client
    // epayment-specific fields
}

// Implements usecase.PaymentGateway interface
func (g *Gateway) GetAuthToken(ctx context.Context) (string, error) {
    // epayment.kz specific implementation
    // ...
}

func (g *Gateway) GetTerminal() string {
    return g.config.Terminal
}

// ... other interface methods
```

### Dependency Injection

```go
// internal/infrastructure/app/app.go
func NewApp(config *Config) *App {
    // Create concrete gateway
    paymentGateway := epayment.NewGateway(config.Payment)

    // Pass as interface to container
    gatewayServices := &usecase.GatewayServices{
        PaymentGateway: paymentGateway,  // Concrete → Interface
    }

    container := usecase.NewContainer(repos, caches, authSvcs, gatewayServices)
    // ...
}
```

## Consequences

### Positive

✅ **Testable:** Easy to mock gateway in use case tests
```go
// Test with mock gateway
type mockGateway struct{}

func (m *mockGateway) GetAuthToken(ctx context.Context) (string, error) {
    return "mock-token", nil
}

func TestInitiatePayment(t *testing.T) {
    mockGateway := &mockGateway{}
    uc := NewInitiatePaymentUseCase(repo, service, mockGateway)
    // Test without real gateway
}
```

✅ **Flexible:** Easy to switch gateways
```go
// Switch to different gateway
paymentGateway := stripe.NewGateway(config)  // New implementation
gatewayServices := &usecase.GatewayServices{
    PaymentGateway: paymentGateway,  // Same interface!
}
```

✅ **Clean dependencies:** Use cases depend on abstraction, not implementation

✅ **Domain-driven:** Interface shaped by business needs, not gateway API

### Negative

❌ **Extra layer:** Need to maintain interface + implementation

❌ **Return type `interface{}`:** CheckPaymentStatus returns generic interface
- **Why:** Different gateways return different response formats
- **Better approach:** Define common response type (future improvement)

## Evolution Path

### Current State (Acceptable)

```go
type PaymentGateway interface {
    CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error)
}
```

**Problem:** `interface{}` is too generic

### Future Improvement

```go
// Define common response type
type PaymentStatusResponse struct {
    Status      string
    Amount      int64
    Currency    string
    CardMask    *string
    ErrorCode   *string
    ErrorMessage *string
}

type PaymentGateway interface {
    CheckPaymentStatus(ctx context.Context, invoiceID string) (*PaymentStatusResponse, error)
}
```

**Benefits:**
- Type-safe
- Gateway-agnostic structure
- Each implementation maps its response to common type

## Interface Design Principles

### 1. Minimal Interface

Only include methods use cases actually need:

```go
// ✅ GOOD: Minimal, focused
type PaymentGateway interface {
    GetAuthToken(ctx context.Context) (string, error)
    CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error)
}

// ❌ BAD: Exposing too much
type PaymentGateway interface {
    GetAuthToken(ctx context.Context) (string, error)
    RefreshToken(ctx context.Context) error  // Not used by use cases
    GetHTTPClient() *http.Client              // Implementation detail
    SetTimeout(duration time.Duration)        // Configuration detail
}
```

### 2. Use Case Driven

Interface methods match use case needs, not gateway API:

```go
// ✅ GOOD: Business operation
GetAuthToken(ctx context.Context) (string, error)

// ❌ BAD: Gateway-specific
ExecuteHTTPPostWithOAuth2(url string, body []byte) ([]byte, error)
```

### 3. Context-Aware

All methods that make external calls should accept `context.Context`:

```go
// ✅ GOOD: Can be cancelled/timeout
GetAuthToken(ctx context.Context) (string, error)

// ❌ BAD: Can't control timeout
GetAuthToken() (string, error)
```

## Current Interface

```go
type PaymentGateway interface {
    // Authentication
    GetAuthToken(ctx context.Context) (string, error)

    // Configuration
    GetTerminal() string
    GetBackLink() string
    GetPostLink() string
    GetWidgetURL() string

    // Operations
    CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error)
}
```

**Why these methods:**
- `GetAuthToken`: Required for every gateway request
- `GetTerminal/BackLink/PostLink/WidgetURL`: Configuration needed by frontend
- `CheckPaymentStatus`: Query payment state

**Missing (intentionally):**
- Payment initiation: Too complex, different per gateway
- Refund: Not all gateways support same refund flow
- Tokenization: Gateway-specific implementation

## Related Decisions

- **ADR 002:** Clean Architecture - Interfaces in use case layer
- **ADR 003:** Domain Services vs Infrastructure - Gateway is infrastructure

## References

- **Interface:** `internal/usecase/container.go` - PaymentGateway interface
- **Implementation:** `internal/adapters/payment/epayment/gateway.go`
- **Usage:** `internal/usecase/paymentops/` - All payment use cases

## Notes for AI Assistants

### Adding Gateway Method

1. **Add to interface:**
```go
// internal/usecase/container.go
type PaymentGateway interface {
    NewMethod(ctx context.Context, param string) (Result, error)
}
```

2. **Implement in adapter:**
```go
// internal/adapters/payment/epayment/gateway.go
func (g *Gateway) NewMethod(ctx context.Context, param string) (Result, error) {
    // Implementation
}
```

3. **Use in use case:**
```go
// internal/usecase/paymentops/some_usecase.go
result, err := uc.gateway.NewMethod(ctx, param)
```

### Testing with Mock

```go
type mockPaymentGateway struct {
    mock.Mock
}

func (m *mockPaymentGateway) GetAuthToken(ctx context.Context) (string, error) {
    args := m.Called(ctx)
    return args.String(0), args.Error(1)
}

func TestUseCase(t *testing.T) {
    mockGateway := new(mockPaymentGateway)
    mockGateway.On("GetAuthToken", mock.Anything).Return("token", nil)

    uc := NewUseCase(repo, service, mockGateway)
    // Test...
}
```

## Revision History

- **2025-10-09:** Initial ADR documenting gateway abstraction pattern
