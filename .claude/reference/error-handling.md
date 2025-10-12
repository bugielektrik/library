# Error Handling Guide

**Last Updated:** 2025-10-10
**Applies to:** All layers (Domain, Use Case, Adapters, Infrastructure)

## Philosophy

**Errors should be:**
- **Actionable** - Clear what went wrong and why
- **Contextual** - Include relevant details for debugging
- **Consistent** - Same patterns across the codebase
- **User-friendly** - Map to appropriate HTTP status codes
- **Traceable** - Preserve error chains for debugging

---

## Error Types

### Custom Domain Errors (`pkg/errors/*.go`)

Use for **business logic failures** and **expected error conditions**.

**Available Errors:**

#### Generic Errors
- `ErrValidation` - Input validation failures (400)
- `ErrInvalidInput` - Invalid input data (400)
- `ErrNotFound` - Resource not found (404)
- `ErrAlreadyExists` - Resource already exists (409)
- `ErrUnauthorized` - Authentication required (401)
- `ErrForbidden` - Access forbidden (403)
- `ErrInternal` - Internal server error (500)
- `ErrDatabase` - Database operation failed (500)
- `ErrCache` - Cache operation failed (500)
- `ErrBusinessRule` - Business rule violation (422)

#### Domain-Specific Errors
- `ErrBookNotFound`, `ErrBookAlreadyExists`, `ErrInvalidBookData`, `ErrInvalidISBN`
- `ErrAuthorNotFound`, `ErrAuthorAlreadyExists`, `ErrInvalidAuthorData`
- `ErrMemberNotFound`, `ErrMemberAlreadyExists`, `ErrInvalidMemberData`, `ErrMembershipExpired`
- `ErrPaymentNotFound`, `ErrPaymentAlreadyProcessed`, `ErrPaymentExpired`, `ErrPaymentGateway`
- `ErrInvalidCredentials`, `ErrInvalidToken`

### Standard Library Errors (`fmt.Errorf`)

Use **only** for:
- Infrastructure/unexpected errors
- External library errors
- System-level failures
- Internal implementation details

---

## Rules by Layer

### 1. Domain Layer (`internal/domain/*`)

**Rule:** Use domain-specific errors for business logic violations.

✅ **DO:**
```go
// Use specific domain errors
if book.Name == nil || *book.Name == "" {
    return errors.ErrInvalidBookData.
        WithDetails("field", "name").
        WithDetails("reason", "required")
}

// Validate business rules
if len(book.Authors) == 0 {
    return errors.ErrInvalidBookData.
        WithDetails("field", "authors").
        WithDetails("reason", "at least one author required")
}

// ISBN validation
if !isValidISBN13(isbn) {
    return errors.ErrInvalidISBN.
        WithDetails("isbn", isbn).
        WithDetails("reason", "invalid checksum")
}
```

❌ **DON'T:**
```go
// Don't use fmt.Errorf for business logic
return fmt.Errorf("book name is required")

// Don't use generic errors without context
return errors.ErrValidation

// Don't return nil errors
if someCondition {
    return nil // This is confusing
}
```

### 2. Use Case Layer (`internal/usecase/*`)

**Rule:** Use generic errors with rich context. Convert repository errors to domain errors.

✅ **DO:**
```go
// Validate request with context
func (r CreateBookRequest) Validate() error {
    if r.Name == "" {
        return errors.ErrValidation.
            WithDetails("field", "name").
            WithDetails("reason", "required")
    }

    if len(r.ISBN) != 10 && len(r.ISBN) != 13 {
        return errors.ErrValidation.
            WithDetails("field", "isbn").
            WithDetails("reason", "invalid format").
            WithDetails("expected", "10 or 13 characters").
            WithDetails("actual", len(r.ISBN))
    }

    return nil
}

// Convert repository errors
book, err := uc.bookRepo.GetByID(ctx, id)
if err != nil {
    if errors.Is(err, store.ErrNotFound) {
        return errors.ErrBookNotFound.WithDetails("book_id", id)
    }
    return errors.ErrDatabase.Wrap(err)
}

// Wrap domain service errors
if err := uc.bookService.ValidateBook(bookEntity); err != nil {
    // Domain error already has context, just return it
    return err
}

// Add context when calling external service
if err := uc.paymentGateway.ChargeCard(ctx, req); err != nil {
    logger.Error("payment provider failed", zap.Error(err))
    return errors.ErrPaymentGateway.
        WithDetails("invoice_id", req.InvoiceID).
        Wrap(err)
}
```

❌ **DON'T:**
```go
// Don't use fmt.Errorf for business logic
if book == nil {
    return fmt.Errorf("book not found: %s", id)
}

// Don't return repository errors directly
book, err := uc.bookRepo.GetByID(ctx, id)
if err != nil {
    return err // Leaks implementation details
}

// Don't lose error context
if err := something(); err != nil {
    return errors.ErrInternal // Lost original error
}
```

### 3. Adapter Layer (`internal/adapters/*`)

**Rule:** Convert external errors to domain errors. Use fmt.Errorf for internal adapter logic.

✅ **DO:**
```go
// HTTP Handlers - Convert to HTTP responses
func (h *BookHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    // ...
    resp, err := h.createBookUC.Execute(ctx, req)
    if err != nil {
        h.RespondError(w, r, err) // Auto-maps error to HTTP status
        return
    }
    h.RespondJSON(w, http.StatusCreated, resp)
}

// Repository - Convert database errors
func (r *PostgresBookRepo) GetByID(ctx context.Context, id string) (book.Book, error) {
    var b book.Book
    err := r.db.GetContext(ctx, &b, query, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return book.Book{}, store.ErrNotFound
        }
        return book.Book{}, fmt.Errorf("querying book by id: %w", err)
    }
    return b, nil
}

// Payment Gateway - Wrap external errors
func (g *EPaymentGateway) ChargeCard(ctx context.Context, req *payment.CardChargeRequest) (*payment.GatewayResponse, error) {
    resp, err := g.client.Post(ctx, "/charge", req)
    if err != nil {
        return nil, fmt.Errorf("payment provider request failed: %w", err)
    }

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("payment provider returned %d: %s", resp.StatusCode, resp.Body)
    }

    return parseResponse(resp), nil
}
```

❌ **DON'T:**
```go
// Don't return database-specific errors
return book.Book{}, err // sql.ErrNoRows leaks to domain

// Don't create new domain errors in adapters
return errors.ErrBookNotFound // Should happen in use case layer

// Don't ignore errors
_ = r.cache.Set(ctx, key, value) // Should at least log
```

### 4. Infrastructure Layer (`internal/infrastructure/*`)

**Rule:** Use fmt.Errorf for configuration, initialization, and system errors.

✅ **DO:**
```go
// Configuration errors
func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("loading config from environment: %w", err)
    }

    if cfg.DatabaseDSN == "" {
        return nil, fmt.Errorf("DATABASE_DSN is required")
    }

    return cfg, nil
}

// Database connection errors
func NewPostgresStore(dsn string) (*Store, error) {
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("connecting to postgres: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("pinging postgres: %w", err)
    }

    return &Store{db: db}, nil
}
```

---

## Patterns and Examples

### Pattern 1: Request Validation

**Always validate in use case request structs:**

```go
type CreateBookRequest struct {
    Name    string   `json:"name"`
    Genre   string   `json:"genre"`
    ISBN    string   `json:"isbn"`
    Authors []string `json:"authors"`
}

func (r CreateBookRequest) Validate() error {
    if r.Name == "" {
        return errors.ErrValidation.
            WithDetails("field", "name").
            WithDetails("reason", "required")
    }

    if r.Genre == "" {
        return errors.ErrValidation.
            WithDetails("field", "genre").
            WithDetails("reason", "required")
    }

    if r.ISBN == "" {
        return errors.ErrValidation.
            WithDetails("field", "isbn").
            WithDetails("reason", "required")
    }

    // Basic format check
    cleanISBN := stripNonAlphanumeric(r.ISBN)
    if len(cleanISBN) != 10 && len(cleanISBN) != 13 {
        return errors.ErrValidation.
            WithDetails("field", "isbn").
            WithDetails("reason", "invalid format").
            WithDetails("expected", "10 or 13 characters").
            WithDetails("actual", len(cleanISBN))
    }

    if len(r.Authors) == 0 {
        return errors.ErrValidation.
            WithDetails("field", "authors").
            WithDetails("reason", "at least one author required")
    }

    return nil
}

// Use in Execute
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
    // Step 1: Validate request
    if err := req.Validate(); err != nil {
        return CreateBookResponse{}, err
    }

    // ... rest of logic
}
```

### Pattern 2: Not Found Errors

**Convert repository not found to domain not found:**

```go
// In use case
book, err := uc.bookRepo.GetByID(ctx, req.BookID)
if err != nil {
    if errors.Is(err, store.ErrNotFound) {
        return CreateReservationResponse{}, errors.ErrBookNotFound.
            WithDetails("book_id", req.BookID)
    }
    return CreateReservationResponse{}, errors.ErrDatabase.Wrap(err)
}

member, err := uc.memberRepo.GetByID(ctx, req.MemberID)
if err != nil {
    if errors.Is(err, store.ErrNotFound) {
        return CreateReservationResponse{}, errors.ErrMemberNotFound.
            WithDetails("member_id", req.MemberID)
    }
    return CreateReservationResponse{}, errors.ErrDatabase.Wrap(err)
}
```

### Pattern 3: Business Rule Violations

**Use specific errors with clear context:**

```go
// Check business rules
if !savedCard.CanBeUsed() {
    return errors.ErrValidation.
        WithDetails("field", "saved_card_id").
        WithDetails("reason", "card is inactive or expired").
        WithDetails("is_active", savedCard.IsActive).
        WithDetails("is_expired", savedCard.IsExpired()).
        WithDetails("expiry_month", savedCard.ExpiryMonth).
        WithDetails("expiry_year", savedCard.ExpiryYear)
}

// Check authorization
if savedCard.MemberID != req.MemberID {
    return errors.ErrForbidden.
        WithDetails("saved_card_id", savedCard.ID).
        WithDetails("card_owner", savedCard.MemberID).
        WithDetails("requesting_member", req.MemberID)
}

// Check state transitions
if err := uc.paymentService.ValidateStatusTransition(payment.Status, newStatus); err != nil {
    return errors.ErrInvalidPaymentStatus.
        WithDetails("current_status", payment.Status).
        WithDetails("new_status", newStatus).
        Wrap(err)
}
```

### Pattern 4: External Service Errors

**Wrap and add context:**

```go
// Payment provider
gatewayResp, err := uc.paymentGateway.ChargeCard(ctx, chargeReq)
if err != nil {
    logger.Error("payment provider failed",
        zap.String("invoice_id", chargeReq.InvoiceID),
        zap.Error(err))

    // Update payment status
    _ = uc.paymentRepo.UpdateStatus(ctx, paymentEntity.ID, payment.StatusFailed)

    return PayWithSavedCardResponse{}, errors.ErrPaymentGateway.
        WithDetails("invoice_id", chargeReq.InvoiceID).
        WithDetails("amount", chargeReq.Amount).
        WithDetails("currency", chargeReq.Currency).
        Wrap(err)
}

// Email service
if err := uc.emailService.SendWelcomeEmail(ctx, member.Email, member.FullName); err != nil {
    // Non-critical error, just log
    logger.Warn("failed to send welcome email",
        zap.String("member_id", member.ID),
        zap.String("email", member.Email),
        zap.Error(err))
    // Don't return error - email failure shouldn't fail registration
}
```

### Pattern 5: Multiple Operations

**Add context to each step:**

```go
func (uc *PayWithSavedCardUseCase) Execute(ctx context.Context, req PayWithSavedCardRequest) (PayWithSavedCardResponse, error) {
    logger := logutil.UseCaseLogger(ctx, "pay_with_saved_card", ...)

    // Step 1: Validate and retrieve saved card
    savedCard, err := uc.validateSavedCard(ctx, req.SavedCardID, req.MemberID, logger)
    if err != nil {
        // Error already has context from validateSavedCard
        return PayWithSavedCardResponse{}, err
    }

    // Step 2: Generate invoice and create payment record
    invoiceID := uc.paymentService.GenerateInvoiceID(req.MemberID, req.PaymentType)
    paymentEntity, err := uc.createPaymentRecord(ctx, req, invoiceID, savedCard.CardMask, logger)
    if err != nil {
        // Error already has context from createPaymentRecord
        return PayWithSavedCardResponse{}, err
    }

    // Step 3: Charge card via provider
    paymentEntity, err = uc.chargeCardViaGateway(ctx, paymentEntity, req, savedCard.CardToken, logger)
    if err != nil {
        // Error already has context from chargeCardViaGateway
        return PayWithSavedCardResponse{}, err
    }

    // Step 4: Update card last used (best effort)
    uc.updateCardLastUsed(ctx, savedCard, logger)

    return PayWithSavedCardResponse{...}, nil
}
```

---

## Error Context Guidelines

### What to Include

**Always include:**
- Field name (for validation errors)
- Entity ID (for not found errors)
- Expected vs actual values (for comparison errors)
- Related entity IDs (for relationship errors)

**Example:**
```go
return errors.ErrValidation.
    WithDetails("field", "amount").
    WithDetails("reason", "below minimum").
    WithDetails("minimum", payment.MinPaymentAmount).
    WithDetails("actual", req.Amount).
    WithDetails("currency", req.Currency)
```

**Never include:**
- Sensitive data (passwords, tokens, card numbers)
- Full credit card numbers (use masked: `****1234`)
- Personal information (use IDs, not names/emails)

### Logging with Errors

**Use structured logging alongside errors:**

```go
if err != nil {
    logger.Error("failed to process payment",
        zap.String("payment_id", paymentID),
        zap.String("member_id", memberID),
        zap.Int64("amount", amount),
        zap.Error(err))
    return errors.ErrPaymentGateway.WithDetails("payment_id", paymentID).Wrap(err)
}
```

---

## Testing Error Handling

### Table-Driven Tests

```go
func TestCreateBook_Validation(t *testing.T) {
    tests := []struct {
        name          string
        request       CreateBookRequest
        expectedError *errors.Error
        expectedField string
    }{
        {
            name: "missing name",
            request: CreateBookRequest{
                Genre:   "Technology",
                ISBN:    "9780132350884",
                Authors: []string{"author-1"},
            },
            expectedError: errors.ErrValidation,
            expectedField: "name",
        },
        {
            name: "invalid ISBN format",
            request: CreateBookRequest{
                Name:    "Test Book",
                Genre:   "Technology",
                ISBN:    "invalid",
                Authors: []string{"author-1"},
            },
            expectedError: errors.ErrValidation,
            expectedField: "isbn",
        },
        {
            name: "no authors",
            request: CreateBookRequest{
                Name:    "Test Book",
                Genre:   "Technology",
                ISBN:    "9780132350884",
                Authors: []string{},
            },
            expectedError: errors.ErrValidation,
            expectedField: "authors",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.request.Validate()

            // Check error type
            assert.True(t, errors.Is(err, tt.expectedError))

            // Check error details
            var domainErr *errors.Error
            if errors.As(err, &domainErr) {
                assert.Equal(t, tt.expectedField, domainErr.Details["field"])
            }
        })
    }
}
```

---

## Migration Checklist

When updating existing code to follow this guide:

### 1. Request Validation
- [ ] Add `Validate()` method to all request structs
- [ ] Use `errors.ErrValidation` with field details
- [ ] Call validation first in `Execute()`

### 2. Repository Error Handling
- [ ] Convert `store.ErrNotFound` to domain-specific not found errors
- [ ] Wrap database errors with `errors.ErrDatabase.Wrap(err)`
- [ ] Add entity ID to error details

### 3. Business Logic Errors
- [ ] Use domain-specific errors when available
- [ ] Use generic errors with rich context otherwise
- [ ] Include all relevant details for debugging

### 4. External Service Errors
- [ ] Wrap external errors with appropriate domain error
- [ ] Add context (IDs, amounts, etc.)
- [ ] Log errors with structured fields
- [ ] Update entity status on failure when needed

### 5. Infrastructure Errors
- [ ] Use `fmt.Errorf` for config/startup errors
- [ ] Include clear error messages
- [ ] Preserve error chain with `%w`

---

## Quick Reference

```go
// ✅ Validation error
errors.ErrValidation.WithDetails("field", "email").WithDetails("reason", "required")

// ✅ Not found error
errors.ErrBookNotFound.WithDetails("book_id", id)

// ✅ Already exists error
errors.ErrBookAlreadyExists.WithDetails("isbn", isbn)

// ✅ Business rule error
errors.ErrBusinessRule.WithDetails("rule", "max_reservations").WithDetails("max", 5)

// ✅ Authorization error
errors.ErrForbidden.WithDetails("resource", "payment").WithDetails("owner", ownerID)

// ✅ Database error
errors.ErrDatabase.Wrap(err)

// ✅ External service error
errors.ErrPaymentGateway.WithDetails("invoice_id", id).Wrap(err)

// ✅ Infrastructure error (config, startup)
fmt.Errorf("loading config: %w", err)
```

---

## Additional Resources

- **Error Package:** `pkg/errors/errors.go`
- **Domain Errors:** `pkg/errors/domain.go`
- **Error Handling Examples:** `pkg/errors/example_test.go`
- **Handler Error Responses:** `internal/infrastructure/pkg/handlers/base.go`
- **Common Mistakes:** `.claude/COMMON-MISTAKES.md`

---

**Last Updated:** 2025-10-10
**Questions?** Open an issue or ask in #engineering channel
