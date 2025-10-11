# ADR 001: Use Case Packages Use "ops" Suffix

**Status:** Accepted

**Date:** 2025-10-09

**Context:**

When importing both domain and use case packages, naming conflicts occur because Go doesn't allow importing two packages with the same name into the same file:

```go
// PROBLEM: Import conflict
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/book"  // ERROR: redeclared as imported package name
)

// This forces ugly workarounds:
import (
    "library-service/internal/domain/book"
    bookUC "library-service/internal/usecase/book"  // Import alias required
)
```

This issue arises frequently because:
- Use cases orchestrate domain entities
- Handlers need both domain types (for entities) and use case types (for operations)
- Import aliases make code less readable and non-idiomatic

## Decision

All use case packages use the **"ops" suffix** (operations):

```
internal/
├── domain/
│   ├── book/          # Package: book
│   ├── member/        # Package: member
│   └── payment/       # Package: payment
└── usecase/
    ├── bookops/       # Package: bookops  ← "ops" suffix
    ├── memberops/     # Package: memberops
    └── paymentops/    # Package: paymentops
```

## Consequences

### Positive

✅ **No import aliases needed:**
```go
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"  // No conflict!
)

// Clean, idiomatic code:
bookEntity := book.New(...)
useCase := bookops.NewCreateBookUseCase(...)
```

✅ **Clear semantic distinction:**
- `book` package = domain entities and business rules
- `bookops` package = operations on books (use cases)

✅ **Idiomatic Go:** No renaming imports, follows Go conventions

✅ **Consistent across codebase:** All use case packages follow same pattern

### Negative

❌ **Directory names don't match domain exactly:**
- Domain: `internal/domain/book`
- Use Case: `internal/usecase/bookops` (not `book`)

❌ **Package names slightly longer:**
- `bookops.CreateBookUseCase` vs `book.CreateBookUseCase`

❌ **New developers need to learn the convention**

## Alternatives Considered

### 1. Import Aliases (Rejected)
```go
import (
    "library-service/internal/domain/book"
    bookUC "library-service/internal/usecase/book"  // Alias
)
```

**Why rejected:**
- Not idiomatic Go (aliases should be rare)
- Inconsistent (some files need aliases, some don't)
- Makes code harder to read and maintain

### 2. Different Directory Structure (Rejected)
```
internal/
├── book/
│   ├── domain/      # Package: domain
│   └── usecase/     # Package: usecase
```

**Why rejected:**
- Breaks Clean Architecture layer separation
- Makes dependency management harder
- Violates single responsibility at directory level

### 3. No Suffix, Just Live With Aliases (Rejected)
**Why rejected:**
- Leads to inconsistent codebase
- Import aliases everywhere reduce code quality
- Not scalable as project grows

## Examples

### Handler Using Both Packages

```go
package book

import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"
)

type BookHandler struct {
    useCases struct {
        CreateBook *bookops.CreateBookUseCase  // Use case
        GetBook    *bookops.GetBookUseCase
    }
}

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    // Use domain entity
    bookEntity := book.New(name, isbn, authors)

    // Call use case
    result, err := h.useCases.CreateBook.Execute(ctx, req)
}
```

### Use Case Importing Domain

```go
package bookops

import (
    "library-service/internal/domain/book"
)

type CreateBookUseCase struct {
    repo    book.Repository  // Domain interface
    service *book.Service    // Domain service
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*CreateBookResponse, error) {
    // Validate with domain service
    if err := uc.service.ValidateISBN(req.ISBN); err != nil {
        return nil, err
    }

    // Create domain entity
    bookEntity := book.New(req.Name, req.ISBN, req.Authors)

    // Persist
    id, err := uc.repo.Create(ctx, bookEntity)
    // ...
}
```

## Related Decisions

- **ADR 002:** Clean Architecture Boundaries - Why use cases are separate layer
- **ADR 003:** Domain Services vs Infrastructure Services - Where services belong

## References

- **Implementation:** `internal/usecase/bookops/`, `internal/usecase/paymentops/`
- **Documentation:** `.claude/CLAUDE.md` - Package naming section
- **Discussion:** Initial implementation in commit fa693e6

## Notes for AI Assistants

When creating new features:
1. ✅ Use "ops" suffix for use case packages
2. ✅ Keep domain packages without suffix
3. ✅ No import aliases needed
4. ❌ Don't create packages named just "usecase" or "usecases"

## Revision History

- **2025-10-09:** Initial ADR documenting existing pattern
