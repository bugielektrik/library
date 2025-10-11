# ADR 002: Clean Architecture Layer Boundaries

**Status:** Accepted

**Date:** 2025-10-09

**Context:**

Large codebases often suffer from:
- **Circular dependencies:** Module A imports B, B imports C, C imports A
- **Tight coupling:** Changes in one area ripple through the entire system
- **Hard to test:** Business logic mixed with framework/database code
- **Difficult to change:** Swapping databases or frameworks requires massive refactoring

Traditional MVC or layered architectures don't enforce strict boundaries, leading to:
```go
// BAD: Everything mixed together
type BookController struct {
    db *sql.DB  // Direct database dependency
}

func (c *BookController) CreateBook(w http.ResponseWriter, r *http.Request) {
    // HTTP parsing
    var req CreateBookRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Business logic
    if len(req.ISBN) != 17 {
        http.Error(w, "invalid ISBN", 400)
        return
    }

    // Direct SQL
    _, err := c.db.Exec("INSERT INTO books ...")
    // ...
}
```

## Decision

Adopt **Clean Architecture** with strict dependency rules:

```
┌─────────────────────────────────────────────┐
│ Domain (Core Business Logic)                │  ← Innermost: ZERO dependencies
│   - Entities                                │
│   - Business Rules                          │
│   - Repository Interfaces                   │
└─────────────────────────────────────────────┘
                 ↑
                 │ depends on (uses interfaces)
                 │
┌─────────────────────────────────────────────┐
│ Use Case (Application Logic)                │  ← Orchestrates domain
│   - Use Cases                               │
│   - Application Services                    │
└─────────────────────────────────────────────┘
                 ↑
                 │ depends on
                 │
┌─────────────────────────────────────────────┐
│ Adapters (Interface Layer)                  │  ← Implements interfaces
│   - HTTP Handlers                           │
│   - Repository Implementations              │
│   - DTOs, Serialization                     │
└─────────────────────────────────────────────┘
                 ↑
                 │ depends on
                 │
┌─────────────────────────────────────────────┐
│ Infrastructure (External Concerns)          │  ← Outermost: frameworks
│   - Database Connections                    │
│   - Configuration                           │
│   - Logging, Monitoring                     │
└─────────────────────────────────────────────┘
```

**THE DEPENDENCY RULE:**

Dependencies point **INWARD ONLY**. Outer layers depend on inner layers, never the reverse.

## Implementation

### Layer 1: Domain (Core)

**Location:** `internal/domain/`

**Purpose:** Pure business logic, zero external dependencies

**Contains:**
- Entities (Book, Member, Payment)
- Domain Services (business rules)
- Repository Interfaces (define what we need, not how it's implemented)

**Rules:**
- ✅ Can import: Standard library only
- ❌ Cannot import: Use cases, adapters, infrastructure, external libraries

```go
// internal/domain/book/book.go
package book

// Pure entity - no dependencies
type Book struct {
    ID      string
    Title   string
    ISBN    string
    Authors []string
}

// Domain service - pure business logic
type Service struct{}

func (s *Service) ValidateISBN(isbn string) error {
    // Business rules only, no external dependencies
    if len(isbn) != 17 {
        return errors.New("ISBN must be 17 characters")
    }
    return nil
}

// Repository interface - defines what we need
type Repository interface {
    Create(ctx context.Context, book Book) (string, error)
    GetByID(ctx context.Context, id string) (Book, error)
}
```

### Layer 2: Use Case (Application Logic)

**Location:** `internal/usecase/`

**Purpose:** Orchestrate domain entities and services

**Contains:**
- Use Cases (CreateBook, ProcessPayment)
- Application services (coordinate multiple domains)

**Rules:**
- ✅ Can import: Domain layer
- ❌ Cannot import: Adapters, infrastructure (uses interfaces)

```go
// internal/usecase/bookops/create_book.go
package bookops

import "library-service/internal/domain/book"

type CreateBookUseCase struct {
    repo    book.Repository  // Interface from domain
    service *book.Service    // Domain service
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*CreateBookResponse, error) {
    // Validate using domain service
    if err := uc.service.ValidateISBN(req.ISBN); err != nil {
        return nil, err
    }

    // Create domain entity
    bookEntity := book.Book{
        Title:   req.Title,
        ISBN:    req.ISBN,
        Authors: req.Authors,
    }

    // Persist via repository interface
    id, err := uc.repo.Create(ctx, bookEntity)
    return &CreateBookResponse{BookID: id}, err
}
```

### Layer 3: Adapters (Interface Implementations)

**Location:** `internal/adapters/`

**Purpose:** Implement domain interfaces, handle external formats

**Contains:**
- HTTP handlers (receive HTTP, call use cases)
- Repository implementations (PostgreSQL, MongoDB)
- DTOs (HTTP JSON ↔ domain entities)

**Rules:**
- ✅ Can import: Domain, use cases
- ✅ Can import: External libraries (database drivers, HTTP frameworks)

```go
// internal/adapters/http/handlers/book/crud.go
package book

import (
    "library-service/internal/usecase/bookops"
    "library-service/internal/adapters/http/dto"
)

type BookHandler struct {
    createBookUC *bookops.CreateBookUseCase
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Decode HTTP request → DTO
    var req dto.CreateBookRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Call use case
    result, err := h.createBookUC.Execute(r.Context(), bookops.CreateBookRequest{
        Title:   req.Title,
        ISBN:    req.ISBN,
        Authors: req.Authors,
    })

    // 3. Encode result → HTTP response
    json.NewEncoder(w).Encode(result)
}
```

```go
// internal/adapters/repository/postgres/book.go
package postgres

import (
    "library-service/internal/domain/book"
    "github.com/jmoiron/sqlx"
)

type BookRepository struct {
    db *sqlx.DB  // External dependency OK in adapter
}

// Implements book.Repository interface
func (r *BookRepository) Create(ctx context.Context, book book.Book) (string, error) {
    query := "INSERT INTO books (id, title, isbn) VALUES ($1, $2, $3)"
    _, err := r.db.ExecContext(ctx, query, book.ID, book.Title, book.ISBN)
    return book.ID, err
}
```

### Layer 4: Infrastructure (External Concerns)

**Location:** `internal/infrastructure/`, `cmd/`

**Purpose:** Bootstrap application, configure external systems

**Contains:**
- Database connections
- Configuration loading
- Dependency injection wiring
- Server setup

```go
// internal/infrastructure/app/app.go
package app

func NewApp() *App {
    // 1. Infrastructure: DB connection
    db := connectPostgreSQL(config.DatabaseURL)

    // 2. Adapters: Repository implementations
    bookRepo := postgres.NewBookRepository(db)

    // 3. Domain: Services
    bookService := book.NewService()

    // 4. Use Cases
    createBookUC := bookops.NewCreateBookUseCase(bookRepo, bookService)

    // 5. Handlers
    bookHandler := handlers.NewBookHandler(createBookUC)

    return &App{Handler: bookHandler}
}
```

## Consequences

### Positive

✅ **Testable:** Mock external dependencies easily
```go
// Test use case without database
mockRepo := mocks.NewMockBookRepository()
uc := bookops.NewCreateBookUseCase(mockRepo, bookService)
```

✅ **Flexible:** Swap implementations without changing business logic
```go
// Easy to change: PostgreSQL → MongoDB
bookRepo := mongodb.NewBookRepository(mongoClient)  // Same interface!
```

✅ **Clear boundaries:** Each layer has single responsibility

✅ **Independent of frameworks:** Domain logic doesn't depend on Chi, GORM, etc.

✅ **Maintainable:** Changes in one layer don't cascade

### Negative

❌ **More files:** Each layer requires separate files
❌ **Boilerplate:** Repository interfaces + implementations
❌ **Learning curve:** Team needs to understand architecture

## Rules Enforcement

### Compile-Time Checks

```bash
# Check for illegal dependencies
go list -json ./internal/domain/... | jq -r '.Deps[]' | grep -E '(usecase|adapters|infrastructure)'
# Should return nothing
```

### Code Review Checklist

- [ ] Domain layer has NO imports from outer layers
- [ ] Use cases only import domain layer
- [ ] Repository implementations in adapters, not domain
- [ ] HTTP/Database logic only in adapters
- [ ] Configuration only in infrastructure

## Examples of Violations

### ❌ BAD: Domain importing use case

```go
// internal/domain/book/book.go
package book

import "library-service/internal/usecase/bookops"  // VIOLATION!

type Book struct {
    // ...
}
```

**Why bad:** Domain depends on application layer (breaks dependency rule)

### ❌ BAD: Domain with database dependency

```go
// internal/domain/book/repository.go
package book

import "github.com/jmoiron/sqlx"  // VIOLATION!

type Repository struct {
    db *sqlx.DB  // External library in domain
}
```

**Why bad:** Domain coupled to database implementation

### ✅ GOOD: Domain defines interface, adapter implements

```go
// internal/domain/book/repository.go
package book

type Repository interface {
    Create(ctx context.Context, book Book) (string, error)
}

// internal/adapters/repository/postgres/book.go
package postgres

type BookRepository struct {
    db *sqlx.DB  // OK here
}

func (r *BookRepository) Create(ctx context.Context, book book.Book) (string, error) {
    // Implementation
}
```

## Related Decisions

- **ADR 001:** Use Case "ops" Suffix - How we organize use cases
- **ADR 003:** Domain Services vs Infrastructure Services - Where to create services

## References

- **Book:** "Clean Architecture" by Robert C. Martin
- **Implementation:** `internal/` directory structure
- **Documentation:** `.claude/architecture.md`

## Notes for AI Assistants

When adding features:
1. ✅ Start with domain layer (entity + interface)
2. ✅ Add use case (orchestration)
3. ✅ Implement adapters (HTTP, database)
4. ✅ Wire in infrastructure

Never:
1. ❌ Put business logic in HTTP handlers
2. ❌ Import adapters into domain
3. ❌ Put SQL queries in use cases

## Revision History

- **2025-10-09:** Initial ADR documenting existing architecture
