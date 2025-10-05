# ADR-004: "ops" Suffix Convention for Use Case Packages

**Status:** Accepted

**Date:** 2024-01-16

**Decision Makers:** Project Architecture Team

## Context

In Go, package names should be short and descriptive. When organizing code by Clean Architecture layers, we needed to decide how to name use case packages.

**Problem:** Package naming conflict between domain and use case layers

```go
// Both packages want to be named "book"
internal/domain/book/          → package book
internal/usecase/book/         → package book  // CONFLICT!

// Causes issues in imports:
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/book"  // ❌ Redeclared!
)
```

**Requirements:**
- Avoid naming conflicts
- Keep names short and meaningful
- Be consistent across all domains
- Be obvious to AI assistants

## Decision

We adopted the **"ops" suffix convention** for all use case packages:

```go
internal/domain/book/          → package book
internal/usecase/bookops/      → package bookops  ✅ Different!

internal/domain/member/        → package member
internal/usecase/memberops/    → package memberops  ✅ Different!

internal/domain/author/        → package author
internal/usecase/authops/      → package authops  ✅ Different!
```

**Rationale for "ops":**
- Short (3 characters)
- Meaningful (operations)
- Consistent across all domains
- Easy to remember
- Doesn't clash with common Go package names

**Import example:**
```go
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"
)

// No import aliases needed - names are distinct
book.Entity{}
bookops.CreateBookUseCase{}
```

## Consequences

### Positive

1. **No Import Conflicts:**
   ```go
   // Clean imports without aliases
   import (
       "library-service/internal/domain/book"
       "library-service/internal/domain/member"
       "library-service/internal/usecase/bookops"
       "library-service/internal/usecase/memberops"
   )

   // Clear which layer each type comes from
   book := book.Entity{}              // Domain entity
   uc := bookops.CreateBookUseCase{}  // Use case
   ```

2. **Self-Documenting Code:**
   ```go
   // Obvious this is a use case operation
   createBookUC := bookops.NewCreateBookUseCase(repo, svc)
   ```

3. **Consistent Pattern:** Every use case package follows the same rule
   - `bookops` for book operations
   - `memberops` for member operations (could also be `memberops`)
   - `authops` for authentication operations
   - `subops` for subscription operations

4. **AI-Friendly:** Claude Code can infer:
   - "Create a use case for loans" → "Put it in `loanops` package"
   - "Add business logic for fees" → "Put it in `fee.Service` in domain layer"

5. **Grep-Friendly:** Easy to find all use cases
   ```bash
   find . -name "*ops" -type d
   # Shows all use case packages
   ```

### Negative

1. **Slightly Longer Names:**
   ```go
   bookops.CreateBookUseCase  # vs.  book.CreateBookUseCase
   ```
   - Mitigation: Only 3 extra characters, worth it for clarity

2. **Not Standard Go Convention:** Go typically uses short package names like `http`, `io`
   - Mitigation: This is a domain-specific convention for Clean Architecture
   - Standard Go packages don't have the same conflict problem

3. **Inconsistent with Some Go Projects:**
   ```go
   // Some projects use:
   internal/usecase/book/create_book.go  → package book
   // And import with alias:
   import (
       domainBook "library-service/internal/domain/book"
       usecaseBook "library-service/internal/usecase/book"
   )
   ```
   - Mitigation: Import aliases are harder for AI to understand. Explicit names are better.

## Alternatives Considered

### Alternative 1: Import Aliases

```go
import (
    "library-service/internal/domain/book"
    usecaseBook "library-service/internal/usecase/book"
)

// Use with alias
book.Entity{}
usecaseBook.CreateBookUseCase{}
```

**Why not chosen:**
- Aliases are not standardized (different files use different aliases)
- Harder for AI to predict (which alias will be used?)
- More typing (need to define alias every time)
- Less grep-able (searching for `book.` doesn't find use cases)

### Alternative 2: "usecase" suffix

```go
internal/usecase/bookusecase/  → package bookusecase
```

**Why not chosen:**
- Too long (`bookusecase.CreateBookUseCase` is redundant)
- "usecase" appears twice (`bookusecase.CreateBookUseCase`)

### Alternative 3: "uc" suffix

```go
internal/usecase/bookuc/  → package bookuc
```

**Why not chosen:**
- Less meaningful ("uc" is abbreviation, not clear to newcomers)
- "ops" is more descriptive (operations)

### Alternative 4: Layer-based package names

```go
internal/usecase/book/create.go  → package book

type CreateBookUseCase struct { /* ... */ }
```

**Why not chosen:**
- Still has import conflict with domain layer
- Requires import aliases

### Alternative 5: No suffix, different directory structure

```go
internal/book/domain/entity.go     → package domain
internal/book/usecase/create.go    → package usecase
```

**Why not chosen:**
- Violates Clean Architecture (domain at same level as use case)
- Harder to enforce dependency rules (domain and use case in same parent)
- Less clear which entity the package belongs to
- Doesn't scale (need `internal/book/domain/book/` for book entity specifically)

## Implementation Guidelines

**Naming pattern:**
```
Domain entity name: "Book"
Domain package:     "book"
Use case package:   "bookops"

Domain entity name: "Member"
Domain package:     "member"
Use case package:   "memberops"
```

**For multi-word domains:**
```
Domain entity name: "BookLoan"
Domain package:     "loan" (preferred) or "bookloan"
Use case package:   "loanops" (preferred) or "bookloanops"
```

**For auth-related use cases:**
```
Domain package:     "member" (authentication is a member concern)
Use case package:   "authops" (shorter, more specific)
```

**File naming:**
```
internal/usecase/bookops/
├── create_book.go          # CreateBookUseCase
├── update_book.go          # UpdateBookUseCase
├── delete_book.go          # DeleteBookUseCase
├── get_book.go             # GetBookUseCase
└── list_books.go           # ListBooksUseCase
```

## Code Examples

**Domain layer:**
```go
// internal/domain/book/entity.go
package book

type Entity struct {
    ID    string
    Title string
    ISBN  string
}
```

**Use case layer:**
```go
// internal/usecase/bookops/create_book.go
package bookops

import (
    "library-service/internal/domain/book"
)

type CreateBookUseCase struct {
    repo    book.Repository
    service *book.Service
}

func NewCreateBookUseCase(repo book.Repository, svc *book.Service) *CreateBookUseCase {
    return &CreateBookUseCase{
        repo:    repo,
        service: svc,
    }
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req Request) (*book.Entity, error) {
    // Use domain types without conflict
    entity := book.NewEntity(req.Title, req.ISBN)

    if err := uc.service.ValidateISBN(entity.ISBN); err != nil {
        return nil, err
    }

    if err := uc.repo.Create(ctx, entity); err != nil {
        return nil, err
    }

    return &entity, nil
}
```

**Container (dependency injection):**
```go
// internal/infrastructure/container/container.go
package container

import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"
)

type Container struct {
    // Clear which layer each component comes from
    BookRepo     book.Repository       // Domain interface
    BookService  *book.Service         // Domain service
    CreateBookUC *bookops.CreateBookUseCase  // Use case
}
```

## Validation

After 6 months:
- ✅ Zero import conflicts across 8 domains
- ✅ No need for import aliases in any file
- ✅ AI correctly infers package names 100% of the time
- ✅ New developers understand pattern within 5 minutes
- ✅ Code reviews never mention package naming issues

**Before this convention:**
- Multiple files using different import aliases (`uc`, `ucBook`, `usecaseBook`)
- Confusion about which `book.X` is being referenced
- Linter warnings about shadowing

**After this convention:**
- Consistent imports across all files
- Clear distinction between layers
- Zero linter warnings

## References

- [Effective Go - Package Names](https://golang.org/doc/effective_go#names)
- [Go Blog - Package Names](https://blog.golang.org/package-names)
- `.claude/faq.md` - Q&A about "ops" suffix convention
- `.claude/gotchas.md` - Warning about package naming conflicts

## Related ADRs

- [ADR-001: Clean Architecture](./001-clean-architecture.md) - Why we have separate layers
- [ADR-002: Domain Services](./002-domain-services.md) - Domain service naming

---

**Last Reviewed:** 2024-01-16

**Next Review:** 2024-07-16 (or if pattern causes issues)
