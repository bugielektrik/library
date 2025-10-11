# Use Case Pattern Standards

**Date:** October 11, 2025
**Status:** ✅ **ACTIVE - Unified Pattern Defined**

## Overview

This document defines the unified code patterns for all use cases in the Library Management System. Following these standards ensures consistency, maintainability, and predictability across the use case layer.

---

## 🎯 Core Principles

1. **Consistency** - Same patterns across all use cases
2. **Predictability** - Know what to expect in any use case
3. **Clarity** - Self-documenting code with clear intent
4. **Testability** - Easy to mock and test
5. **Separation of Concerns** - Use cases orchestrate, domains validate

---

## 📋 Use Case Structure Template

### File Organization

Each use case should be in its own file:

```
internal/usecase/{domain}ops/
├── create_{entity}.go
├── get_{entity}.go
├── update_{entity}.go
├── delete_{entity}.go
├── list_{entity}.go
└── {custom_operation}_{entity}.go
```

**Naming Convention:**
- File: `{action}_{entity}.go` (snake_case)
- Package: `{domain}ops` (e.g., `bookops`, `paymentops`)

---

## 🏗️ Code Pattern

### 1. Request DTO

**Pattern:**
```go
// {Action}{Entity}Request represents the input for {action description}
type {Action}{Entity}Request struct {
    // Required fields
    FieldName string `json:"field_name" validate:"required"`

    // Optional fields (use pointers for optional)
    OptionalField *string `json:"optional_field,omitempty"`
}

// Validate validates the request (if complex validation needed)
func (r {Action}{Entity}Request) Validate() error {
    // Custom validation logic
    return nil
}
```

**Standards:**
- ✅ Naming: `{Action}{Entity}Request`
- ✅ Always add comment describing purpose
- ✅ Use struct tags for JSON marshaling
- ✅ Use pointers for optional fields
- ✅ Add `Validate()` method only if validation is complex
- ✅ Keep validation logic simple - complex rules belong in domain service

**Example:**
```go
// CreateBookRequest represents the input for creating a new book
type CreateBookRequest struct {
    Name    string   `json:"name" validate:"required"`
    Genre   string   `json:"genre" validate:"required"`
    ISBN    string   `json:"isbn" validate:"required"`
    Authors []string `json:"authors" validate:"required,min=1"`
}

func (r CreateBookRequest) Validate() error {
    if len(r.Name) == 0 {
        return errors.ValidationRequired("name")
    }
    return nil
}
```

---

### 2. Response DTO

**Pattern:**
```go
// {Action}{Entity}Response represents the output of {action description}
type {Action}{Entity}Response struct {
    // Response fields
    FieldName string `json:"field_name"`
}
```

**Standards:**
- ✅ Naming: `{Action}{Entity}Response`
- ✅ Always add comment describing purpose
- ✅ Use struct tags for JSON marshaling
- ✅ Return **values**, not pointers (more idiomatic Go)
- ✅ Can embed domain response DTOs if appropriate

**Example:**
```go
// CreateBookResponse represents the output of creating a book
type CreateBookResponse struct {
    book.Response  // Embedded domain response
}

// GetBookResponse represents the output of retrieving a book
type GetBookResponse struct {
    ID      string   `json:"id"`
    Name    string   `json:"name"`
    Genre   string   `json:"genre"`
    ISBN    string   `json:"isbn"`
    Authors []string `json:"authors"`
}
```

---

### 3. Use Case Struct

**Pattern:**
```go
// {Action}{Entity}UseCase handles {description of what it does}.
//
// Architecture Pattern: {pattern description}
// {Additional context or notes}
//
// See Also:
//   - Domain service: internal/domain/{entity}/service.go ({what it does})
//   - Repository: internal/adapters/repository/postgres/{entity}.go
//   - HTTP handler: internal/adapters/http/handlers/{entity}/{action}.go
//   - Test: internal/usecase/{domain}ops/{action}_{entity}_test.go
type {Action}{Entity}UseCase struct {
    {entity}Repo    {entity}.Repository
    {entity}Cache   {entity}.Cache       // Optional
    {entity}Service *{entity}.Service
    // Other dependencies
}
```

**Standards:**
- ✅ Naming: `{Action}{Entity}UseCase`
- ✅ Always add documentation comment
- ✅ Include "See Also" section with cross-references
- ✅ Dependencies: Repository, Service, Cache (if needed)
- ✅ Use interface types for dependencies (Repository, Cache)
- ✅ Use concrete types for services (*Service)

**Example:**
```go
// CreateBookUseCase handles the creation of a new book.
//
// Architecture Pattern: Simple CRUD operation with cache invalidation.
// Demonstrates domain service validation before persistence.
//
// See Also:
//   - Domain service: internal/domain/book/service.go (validation)
//   - Repository: internal/adapters/repository/postgres/book.go
//   - HTTP handler: internal/adapters/http/handlers/book/create.go
//   - Test: internal/usecase/bookops/create_book_test.go
type CreateBookUseCase struct {
    bookRepo    book.Repository
    bookCache   book.Cache
    bookService *book.Service
}
```

---

### 4. Constructor

**Pattern:**
```go
// New{Action}{Entity}UseCase creates a new instance of {Action}{Entity}UseCase
func New{Action}{Entity}UseCase(
    {entity}Repo {entity}.Repository,
    {entity}Cache {entity}.Cache,
    {entity}Service *{entity}.Service,
) *{Action}{Entity}UseCase {
    return &{Action}{Entity}UseCase{
        {entity}Repo:    {entity}Repo,
        {entity}Cache:   {entity}Cache,
        {entity}Service: {entity}Service,
    }
}
```

**Standards:**
- ✅ Naming: `New{Action}{Entity}UseCase`
- ✅ Add comment describing purpose
- ✅ Parameters: All dependencies
- ✅ Return pointer to use case struct
- ✅ Initialize all fields explicitly

**Example:**
```go
// NewCreateBookUseCase creates a new instance of CreateBookUseCase
func NewCreateBookUseCase(
    bookRepo book.Repository,
    bookCache book.Cache,
    bookService *book.Service,
) *CreateBookUseCase {
    return &CreateBookUseCase{
        bookRepo:    bookRepo,
        bookCache:   bookCache,
        bookService: bookService,
    }
}
```

---

### 5. Execute Method

**Pattern:**
```go
// Execute {description of what it does}
func (uc *{Action}{Entity}UseCase) Execute(ctx context.Context, req {Action}{Entity}Request) ({Action}{Entity}Response, error) {
    logger := logutil.UseCaseLogger(ctx, "{domain}", "{action}")

    // 1. Validate request (if Validate method exists)
    if err := req.Validate(); err != nil {
        logger.Warn("validation failed", zap.Error(err))
        return {Action}{Entity}Response{}, err
    }

    // 2. Business logic orchestration
    // - Fetch from repository
    // - Validate with domain service
    // - Create/Update entity
    // - Persist to repository
    // - Update cache

    // 3. Log success
    logger.Info("{action} completed successfully", zap.String("id", id))

    // 4. Return response
    return {Action}{Entity}Response{
        // Response fields
    }, nil
}
```

**Standards:**
- ✅ Signature: `Execute(ctx context.Context, req {Request}) ({Response}, error)`
- ✅ **Always return response DTO** (even for delete/update operations)
- ✅ Response is **value type**, not pointer
- ✅ First action: Create structured logger with `logutil.UseCaseLogger`
- ✅ Log level usage:
  - `Info`: Successful operations
  - `Warn`: Validation errors, business rule violations
  - `Error`: System errors (database, external services)
- ✅ Error handling: Return domain errors, wrap with context
- ✅ Always return empty response on error: `return {Response}{}, err`

**Example:**
```go
// Execute creates a new book in the system
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
    logger := logutil.UseCaseLogger(ctx, "book", "create")

    // Validate request
    if err := req.Validate(); err != nil {
        logger.Warn("validation failed", zap.Error(err))
        return CreateBookResponse{}, err
    }

    // Create book entity
    bookEntity := book.New(book.Request{
        Name:    req.Name,
        Genre:   req.Genre,
        ISBN:    req.ISBN,
        Authors: req.Authors,
    })

    // Validate using domain service
    if err := uc.bookService.Validate(bookEntity); err != nil {
        logger.Warn("domain validation failed", zap.Error(err))
        return CreateBookResponse{}, err
    }

    // Save to repository
    bookID, err := uc.bookRepo.Add(ctx, bookEntity)
    if err != nil {
        logger.Error("failed to create book", zap.Error(err))
        return CreateBookResponse{}, errors.Database("database operation", err)
    }
    bookEntity.ID = bookID

    // Update cache
    if err := uc.bookCache.Set(ctx, bookID, bookEntity); err != nil {
        logger.Warn("failed to update cache", zap.Error(err))
        // Non-critical, continue
    }

    logger.Info("book created successfully", zap.String("id", bookID))

    return CreateBookResponse{
        Response: book.ParseFromBook(bookEntity),
    }, nil
}
```

---

## 🔄 Common Use Case Patterns

### Create Pattern

```go
// 1. Validate request
// 2. Check if entity already exists (if needed)
// 3. Create domain entity
// 4. Validate with domain service
// 5. Persist to repository
// 6. Update cache
// 7. Return response with created entity
```

### Get Pattern

```go
// 1. Validate request (ID required)
// 2. Try to fetch from cache
// 3. If not in cache, fetch from repository
// 4. Update cache with result
// 5. Return response with entity
```

### Update Pattern

```go
// 1. Validate request
// 2. Fetch existing entity from repository
// 3. Apply updates to entity
// 4. Validate with domain service
// 5. Update in repository
// 6. Invalidate/update cache
// 7. Return response with success status
```

### Delete Pattern

```go
// 1. Validate request (ID required)
// 2. Check if entity exists
// 3. Check business rules (can it be deleted?)
// 4. Delete from repository
// 5. Invalidate cache
// 6. Return response with success status
```

### List Pattern

```go
// 1. Validate request (pagination, filters)
// 2. Fetch from repository with filters
// 3. Return response with list of entities
```

---

## ⚠️ Error Handling Standards

### Error Types

**Validation Errors:**
```go
// Use domain errors with details
return errors.NewError(errors.CodeValidation).
    WithField("field_name", "reason").
    Build()

// Or use helpers
return errors.ValidationRequired("field_name")
```

**Not Found Errors:**
```go
// Use domain errors
return errors.NotFoundWithID("entity_name", id)
```

**Database Errors:**
```go
// Wrap with context
return errors.Database("operation description", err)
```

**External Service Errors:**
```go
// Wrap with context
return errors.External("service_name", err)
```

### Error Wrapping

**Always wrap errors with context:**
```go
// ✅ Good
if err != nil {
    logger.Error("failed to create book", zap.Error(err))
    return CreateBookResponse{}, errors.Database("database operation", err)
}

// ❌ Bad - no context
if err != nil {
    return CreateBookResponse{}, err
}
```

---

## 📝 Documentation Standards

### Use Case Comment

**Template:**
```go
// {Action}{Entity}UseCase handles {description}.
//
// Architecture Pattern: {pattern type} - {brief explanation}
// {Additional context or notes}
//
// See Also:
//   - Domain service: internal/domain/{entity}/service.go ({purpose})
//   - Repository: internal/adapters/repository/postgres/{entity}.go
//   - HTTP handler: internal/adapters/http/handlers/{entity}/{action}.go
//   - ADR: .claude/adr/{relevant_adr}.md ({topic})
//   - Test: internal/usecase/{domain}ops/{action}_{entity}_test.go
```

**Architecture Patterns:**
- "Simple CRUD operation" - Basic create/read/update/delete
- "Complex orchestration" - Multiple domain interactions
- "Cross-domain validation" - Checks multiple entities
- "External integration" - Calls external services
- "Async processing" - Background jobs

### Execute Comment

```go
// Execute {describes what the method does in detail}
```

---

## 🧪 Testing Standards

### Test File Structure

```go
package {domain}ops

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "library-service/internal/domain/{entity}"
)

func TestNew{Action}{Entity}UseCase(t *testing.T) {
    // Test constructor
}

func Test{Action}{Entity}UseCase_Execute(t *testing.T) {
    tests := []struct {
        name        string
        request     {Action}{Entity}Request
        setupMocks  func(*MockRepository, *MockService)
        wantErr     bool
        errContains string
    }{
        {
            name: "success case",
            request: {Action}{Entity}Request{
                // Request data
            },
            setupMocks: func(repo *MockRepository, svc *MockService) {
                // Setup mock expectations
            },
            wantErr: false,
        },
        // More test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

---

## 📊 Current State Analysis

### Inconsistencies Found

| Issue | Count | Examples |
|-------|-------|----------|
| **Pointer return types** | 4 | LoginResponse, RegisterResponse, ValidateTokenResponse, RefreshTokenResponse |
| **Missing response DTOs** | 2 | DeleteBookUseCase, UpdateBookUseCase (return only error) |
| **No request parameter** | 2 | ListAuthorsUseCase, ListMembersUseCase |
| **Inconsistent Validate()** | ~50% | Some requests have Validate(), some don't |
| **Mixed error patterns** | All | Some use builder, some use domain errors |

---

## ✅ Refactoring Checklist

### Phase 1: Response Types
- [ ] Change auth use cases to return value types instead of pointers
  - [ ] LoginResponse
  - [ ] RegisterResponse
  - [ ] ValidateTokenResponse
  - [ ] RefreshTokenResponse

### Phase 2: Missing Responses
- [ ] Add response DTOs for:
  - [ ] DeleteBookUseCase
  - [ ] UpdateBookUseCase

### Phase 3: Request Parameters
- [ ] Add request DTOs for:
  - [ ] ListAuthorsUseCase
  - [ ] ListMembersUseCase

### Phase 4: Validation
- [ ] Standardize validation pattern
- [ ] Add Validate() methods where complex validation is needed
- [ ] Remove inline validation where appropriate

### Phase 5: Documentation
- [ ] Add comprehensive comments to all use cases
- [ ] Add "See Also" sections
- [ ] Add architecture pattern descriptions

### Phase 6: Error Handling
- [ ] Standardize error wrapping
- [ ] Use consistent error types

---

## 🎯 Benefits

### Developer Experience
- **Predictability**: Every use case follows the same structure
- **Easier Navigation**: Know exactly where to find things
- **Reduced Cognitive Load**: Same pattern everywhere

### Code Quality
- **Consistency**: All use cases look similar
- **Testability**: Easy to mock and test
- **Maintainability**: Changes are easier to make

### Team Productivity
- **Faster Development**: Less decision fatigue
- **Better Reviews**: Check against standard
- **Easier Onboarding**: New developers see consistent patterns

---

## 📚 Examples

### Complete Example: CreateBookUseCase

See: `internal/usecase/bookops/create_book.go`

### Complete Example: InitiatePaymentUseCase

See: `internal/usecase/paymentops/initiate_payment.go`

### Complete Example: CreateReservationUseCase

See: `internal/usecase/reservationops/create_reservation.go`

---

## 🔗 Related Documents

1. [CODE_PATTERN_STANDARDS.md](./.claude/CODE_PATTERN_STANDARDS.md) - Domain layer patterns
2. [DOMAIN_PATTERN.md](./.claude/DOMAIN_PATTERN.md) - Domain file structure
3. [CODEBASE_PATTERN_REFACTORING.md](./.claude/CODEBASE_PATTERN_REFACTORING.md) - Domain refactoring summary

---

**Generated:** October 11, 2025
**By:** Claude Code (AI-Assisted Pattern Definition)
**Project:** Library Management System
