# Phase 4 Refactoring Plan: Test Modernization & Handler Optimization

## Overview
Phase 4 focuses on applying the improvements from Phases 1-3 to the remaining codebase, particularly test files and HTTP handlers.

## Phase 4A: Test Modernization 🧪

### Current Issues
- **18 test files** still using old mock patterns
- Test helpers scattered across files
- No consistent test data builders
- Missing table-driven tests in some areas

### Tasks

#### 1. Update Test Files to Use Centralized Mocks
**Files to update (18):**
```
internal/usecase/authops/validate_test.go
internal/usecase/authops/login_test.go
internal/usecase/authops/refresh_test.go
internal/usecase/memberops/list_members_test.go
internal/usecase/memberops/get_member_profile_test.go
internal/usecase/subops/subscribe_member_test.go
internal/usecase/paymentops/*.go (10 files)
internal/usecase/reservationops/create_reservation_test.go
```

**Pattern to apply:**
```go
// Old pattern (remove)
type mockMemberRepository struct { ... }

// New pattern (use)
import "library-service/internal/adapters/repository/mocks"
mockRepo := &mocks.MockMemberRepository{}
```

#### 2. Create Test Data Builders
**Create:** `test/builders/`
- `member_builder.go` - Fluent member test data creation
- `payment_builder.go` - Payment test data builder
- `book_builder.go` - Book test data builder
- `reservation_builder.go` - Reservation builder

**Example pattern:**
```go
type MemberBuilder struct {
    member member.Member
}

func NewMemberBuilder() *MemberBuilder {
    return &MemberBuilder{
        member: defaultMember(),
    }
}

func (b *MemberBuilder) WithID(id string) *MemberBuilder {
    b.member.ID = id
    return b
}

func (b *MemberBuilder) Build() member.Member {
    return b.member
}
```

#### 3. Extract Common Test Helpers
**Create:** `test/helpers/`
- `assertions.go` - Common test assertions
- `context.go` - Test context builders
- `fixtures.go` - Common test fixtures

**Benefits:**
- Reduce test code duplication by ~40%
- Consistent test patterns
- Easier test maintenance

## Phase 4B: Handler Optimization 🚀

### Current Issues
- **~30 handlers** using manual validation/error handling
- Repetitive JSON encoding/decoding code
- Inconsistent error responses
- No request/response logging middleware

### Tasks

#### 1. Apply Generic Handler Wrapper
**Handler directories to update:**
```
internal/adapters/http/handlers/auth/
internal/adapters/http/handlers/book/
internal/adapters/http/handlers/member/
internal/adapters/http/handlers/payment/
internal/adapters/http/handlers/reservation/
internal/adapters/http/handlers/savedcard/
```

**Transformation example:**
```go
// Before (40 lines)
func (h *Handler) initiatePayment(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logutil.HandlerLogger(ctx, "payment", "initiate")

    memberID, ok := h.GetMemberID(w, r)
    if !ok { return }

    var req dto.Request
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    if !h.validator.ValidateStruct(w, req) {
        return
    }

    result, err := h.useCase.Execute(ctx, ...)
    if err != nil {
        h.RespondError(w, r, err)
        return
    }

    h.RespondJSON(w, http.StatusOK, result)
}

// After (5 lines)
func (h *Handler) initiatePayment() http.HandlerFunc {
    return httputil.WrapHandler(h.useCase, h.validator, httputil.HandlerOptions{
        RequireAuth: true,
        LogContext: "payment.initiate",
    })
}
```

#### 2. Create Response Transformers
**Create:** `internal/adapters/http/transformers/`
- `payment_transformer.go` - Payment domain to DTO
- `member_transformer.go` - Member transformations
- `book_transformer.go` - Book transformations

**Benefits:**
- Centralized DTO transformations
- Consistent response formats
- Easier to maintain API contracts

#### 3. Add Request Middleware
**Create:** `internal/adapters/http/middleware/`
- `request_id.go` - Add request ID to context
- `request_logger.go` - Log all requests/responses
- `recovery.go` - Panic recovery with stack traces

## Phase 4C: Error & Logging Enhancement 📊

### Current Issues
- Mix of `fmt.Errorf` and custom errors (37 occurrences)
- Inconsistent error wrapping
- No structured logging in some areas
- Missing correlation IDs

### Tasks

#### 1. Standardize Error Creation
**Create:** `pkg/errors/builders.go`
```go
type ErrorBuilder struct {
    err *DomainError
}

func NewError(code string) *ErrorBuilder {
    return &ErrorBuilder{err: &DomainError{Code: code}}
}

func (b *ErrorBuilder) WithMessage(msg string) *ErrorBuilder {
    b.err.Message = msg
    return b
}

func (b *ErrorBuilder) WithDetails(key, value string) *ErrorBuilder {
    b.err.Details[key] = value
    return b
}
```

#### 2. Create Logging Decorators
**Create:** `pkg/logutil/decorators.go`
- Function execution time logging
- Automatic error logging
- Context propagation

**Example:**
```go
func LogExecution(name string, fn func() error) error {
    start := time.Now()
    err := fn()
    logger.Info("executed",
        zap.String("function", name),
        zap.Duration("took", time.Since(start)),
        zap.Error(err),
    )
    return err
}
```

#### 3. Add Correlation IDs
- Generate correlation ID in middleware
- Pass through all layers
- Include in all log entries
- Return in error responses

## Phase 4D: Configuration Management 🔧

### Current Issues
- Direct `os.Getenv` calls scattered
- No configuration validation
- No configuration hot-reload
- Environment-specific configs mixed

### Tasks

#### 1. Create Configuration Types
**Update:** `internal/infrastructure/config/`
```go
type Config struct {
    Server   ServerConfig   `validate:"required"`
    Database DatabaseConfig `validate:"required"`
    Redis    RedisConfig
    JWT      JWTConfig      `validate:"required"`
    Payment  PaymentConfig  `validate:"required"`
}

type ServerConfig struct {
    Port         int           `env:"PORT" default:"8080"`
    ReadTimeout  time.Duration `env:"READ_TIMEOUT" default:"15s"`
    WriteTimeout time.Duration `env:"WRITE_TIMEOUT" default:"15s"`
}
```

#### 2. Add Configuration Validation
- Use struct tags for validation
- Validate on startup
- Provide clear error messages
- Support default values

#### 3. Environment-Specific Configs
**Create config files:**
```
configs/
├── base.yaml        # Common settings
├── development.yaml # Dev overrides
├── staging.yaml     # Staging overrides
└── production.yaml  # Prod overrides
```

## Estimated Impact

### Metrics
| Area | Current | After Phase 4 | Improvement |
|------|---------|---------------|-------------|
| Test duplication | ~2000 lines | ~800 lines | -60% |
| Handler boilerplate | ~30 lines avg | ~8 lines avg | -73% |
| Error consistency | Mixed | Standardized | 100% |
| Config management | Scattered | Centralized | 100% |

### Development Speed
- **Test writing**: 2x faster with builders
- **Handler creation**: 3x faster with wrapper
- **Debugging**: 40% faster with correlation IDs
- **Configuration changes**: No code changes needed

### Code Quality
- **Test coverage**: Easier to achieve 90%+
- **Error handling**: Consistent across codebase
- **Logging**: Structured and searchable
- **Configuration**: Type-safe and validated

## Implementation Order

### Week 1: Test Modernization (4A)
1. Day 1-2: Update test mocks (18 files)
2. Day 3-4: Create test builders
3. Day 5: Extract test helpers

### Week 2: Handler Optimization (4B)
1. Day 1-2: Apply wrapper to auth/member handlers
2. Day 3-4: Apply wrapper to payment/savedcard handlers
3. Day 5: Apply wrapper to book/reservation handlers

### Week 3: Error & Logging (4C)
1. Day 1-2: Standardize error creation
2. Day 3-4: Add logging decorators
3. Day 5: Implement correlation IDs

### Week 4: Configuration (4D)
1. Day 1-2: Create configuration types
2. Day 3-4: Add validation
3. Day 5: Environment-specific configs

## Success Criteria

✅ All test files using centralized mocks
✅ Test builders available for all domains
✅ All handlers using generic wrapper
✅ Consistent error handling throughout
✅ Structured logging with correlation IDs
✅ Centralized, validated configuration
✅ No direct os.Getenv calls
✅ 90%+ test coverage
✅ Handler code reduced by 70%+
✅ Zero test duplication

## Risk Mitigation

### Backward Compatibility
- Keep old patterns working during migration
- Use feature flags for gradual rollout
- Maintain API contracts

### Testing Strategy
- Update tests incrementally
- Run full test suite after each change
- Performance benchmarks before/after

### Rollback Plan
- Git tags at each milestone
- Feature flags for new patterns
- Quick revert procedures documented

---

**Note:** Phase 4 is designed to be implemented incrementally with minimal disruption to ongoing development. Each sub-phase (4A-4D) can be completed independently.