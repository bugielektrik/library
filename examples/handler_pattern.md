# HTTP Handler Pattern

## Overview

Handlers in this codebase follow a consistent pattern using the bounded context structure.

## Structure

```go
package book

import (
    "net/http"
    "library-service/internal/adapters/http/dto"
    "library-service/internal/books/operations"
    "library-service/pkg/httputil"
    "library-service/pkg/logutil"
)

type BookHandler struct {
    handlers.BaseHandler  // Embedded base with RespondJSON, RespondError
    useCases   *usecase.Container
    validator  *middleware.Validator
}

func NewBookHandler(useCases *usecase.Container, validator *middleware.Validator) *BookHandler {
    return &BookHandler{
        BaseHandler: handlers.NewBaseHandler(),
        useCases:    useCases,
        validator:   validator,
    }
}
```

## Handler Method Pattern

All handler methods follow this pattern:

```go
// create is a PRIVATE method (lowercase) that handles book creation
func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Create logger with context
    logger := logutil.HandlerLogger(ctx, "book_handler", "create")

    // 2. Decode request
    var req dto.CreateBookRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    // 3. Validate request
    if !h.validator.ValidateStruct(w, req) {
        return // validator already wrote response
    }

    // 4. Execute use case via grouped container
    result, err := h.useCases.Book.CreateBook.Execute(ctx, operations.CreateBookRequest{
        Name:    req.Name,
        Genre:   req.Genre,
        ISBN:    req.ISBN,
        Authors: req.Authors,
    })
    if err != nil {
        h.RespondError(w, r, err)
        return
    }

    // 5. Convert to DTO
    response := dto.ToBookResponseFromCreate(result)

    // 6. Log and respond
    logger.Info("book created", zap.String("id", response.ID))
    h.RespondJSON(w, http.StatusCreated, response)
}
```

## Key Patterns

### 1. Private Methods
Handler methods are **private** (lowercase) and exposed via routing:

```go
func (h *BookHandler) RegisterRoutes(r chi.Router) {
    r.Post("/books", h.create)      // Private method
    r.Get("/books/{id}", h.get)     // Private method
    r.Put("/books/{id}", h.update)  // Private method
}
```

### 2. Grouped Use Case Access
Use cases accessed via domain groups:

```go
// ✅ CORRECT - Grouped access
h.useCases.Book.CreateBook.Execute(...)
h.useCases.Auth.Login.Execute(...)
h.useCases.Payment.InitiatePayment.Execute(...)

// ❌ WRONG - Flat access (legacy)
h.useCases.CreateBook.Execute(...)
```

### 3. Standard Response Methods
Use BaseHandler methods:

```go
// Success response
h.RespondJSON(w, http.StatusOK, data)

// Error response (handles error types automatically)
h.RespondError(w, r, err)

// Get URL params with validation
id, ok := h.GetURLParam(w, r, "id")
if !ok {
    return // error already written
}
```

### 4. Context Helpers
Always use helpers for context values:

```go
import "library-service/internal/adapters/http/middleware"

// ✅ CORRECT
memberID, ok := middleware.GetMemberIDFromContext(ctx)
if !ok {
    h.RespondError(w, r, errors.ErrUnauthorized)
    return
}

// ❌ WRONG - Direct context access
memberID := ctx.Value("member_id").(string) // Panic risk!
```

## Swagger Annotations

Add Swagger docs before handler methods:

```go
// @Summary Create a new book
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth              // REQUIRED for protected endpoints
// @Param request body dto.CreateBookRequest true "Book data"
// @Success 201 {object} dto.BookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books [post]
func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

## File Organization

Handlers are split by responsibility:

```
internal/books/http/
├── handler.go      # Handler struct, constructor, routes
├── crud.go         # CRUD operations (create, get, update, delete)
├── query.go        # Query operations (list, search, filter)
└── doc.go          # Package documentation
```

## Validation

Use validator adapter consistently:

```go
// Struct validation (most common)
if !h.validator.ValidateStruct(w, req) {
    return // Response already written
}

// Custom validation
if req.Amount <= 0 {
    h.validator.Error(w, "amount", "must be positive")
    return
}
```

## Error Handling

RespondError handles all error types:

```go
// Domain errors → appropriate status codes
errors.ErrNotFound      → 404
errors.ErrAlreadyExists → 409
errors.ErrValidation    → 400
errors.ErrUnauthorized  → 401

// Other errors → 500 Internal Server Error
```

## Complete Example

See actual handlers in:
- `internal/books/http/crud.go` - Book CRUD
- `internal/members/http/auth/handler.go` - Authentication
- `internal/payments/http/payment/initiate.go` - Payment initiation
