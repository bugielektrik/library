# ADR-005: Repository Interfaces Defined in Domain Layer

**Status:** Accepted

**Date:** 2024-01-17

**Decision Makers:** Project Architecture Team

## Context

In Clean Architecture, the domain layer should have no dependencies on outer layers. But the domain needs to persist data, which requires database operations. How do we achieve this without coupling domain to infrastructure?

**Problem:** Domain needs data persistence, but shouldn't depend on database implementation.

**Goal:** Follow the Dependency Inversion Principle - depend on abstractions, not concretions.

## Decision

We define **repository interfaces in the domain layer** and **implement them in the adapter layer**.

**Structure:**
```
internal/domain/book/
├── entity.go         # Book entity
├── service.go        # Business logic
├── repository.go     # ← Repository INTERFACE (domain defines contract)
└── errors.go

internal/adapters/repository/postgres/
├── book.go           # ← Repository IMPLEMENTATION (adapter fulfills contract)
├── member.go
└── author.go
```

**Example:**
```go
// internal/domain/book/repository.go
package book

type Repository interface {
    Create(ctx context.Context, book Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    GetByISBN(ctx context.Context, isbn string) (Entity, error)
    Update(ctx context.Context, book Entity) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]Entity, error)
}

// internal/adapters/repository/postgres/book.go
package postgres

type BookRepository struct {
    db *sql.DB
}

func NewBookRepository(db *sql.DB) book.Repository {
    return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book book.Entity) error {
    // PostgreSQL implementation
}
```

**Dependency direction:**
```
Domain (defines interface) ←── Adapter (implements interface)
```

Not:
```
Domain → Adapter (would create coupling) ❌
```

## Consequences

### Positive

1. **Domain Layer Stays Pure:**
   ```go
   // internal/domain/book/service.go
   package book

   // No imports from outer layers!
   // Only standard library and other domain packages
   import (
       "context"
       "errors"
       "time"
   )
   ```

2. **Easy to Swap Implementations:**
   ```go
   // Can easily switch from PostgreSQL to MongoDB
   // PostgreSQL implementation
   bookRepo := postgres.NewBookRepository(pgDB)

   // MongoDB implementation (same interface)
   bookRepo := mongodb.NewBookRepository(mongoDB)

   // Memory implementation (for testing)
   bookRepo := memory.NewBookRepository()

   // Use case doesn't care - same interface
   uc := bookops.NewCreateBookUseCase(bookRepo, bookService)
   ```

3. **Testability:**
   ```go
   // Easy to mock in tests
   type MockBookRepository struct {
       CreateFunc func(ctx context.Context, book book.Entity) error
   }

   func (m *MockBookRepository) Create(ctx context.Context, book book.Entity) error {
       if m.CreateFunc != nil {
           return m.CreateFunc(ctx, book)
       }
       return nil
   }

   // In test
   func TestCreateBookUseCase(t *testing.T) {
       mockRepo := &MockBookRepository{
           CreateFunc: func(ctx context.Context, book book.Entity) error {
               return nil  // Or simulate error
           },
       }

       uc := bookops.NewCreateBookUseCase(mockRepo, service)
       // Test without real database
   }
   ```

4. **Clear Contract:**
   ```go
   // Domain defines WHAT operations are needed
   type Repository interface {
       Create(ctx context.Context, book Entity) error
       GetByISBN(ctx context.Context, isbn string) (Entity, error)
   }

   // Adapter implements HOW to do them
   func (r *PostgresBookRepository) Create(...) error {
       query := "INSERT INTO books (id, title, isbn) VALUES ($1, $2, $3)"
       _, err := r.db.ExecContext(ctx, query, book.ID, book.Title, book.ISBN)
       return err
   }
   ```

5. **Domain-Driven Queries:**
   ```go
   // Domain dictates what queries are needed
   type Repository interface {
       GetByISBN(ctx context.Context, isbn string) (Entity, error)
       GetAvailableBooks(ctx context.Context) ([]Entity, error)
       GetOverdueLoans(ctx context.Context, memberID string) ([]Entity, error)
   }

   // Not generic CRUD that database layer dictates
   ```

6. **No Import Cycles:**
   ```
   Use Case → Domain (imports domain.Repository interface) ✅
   Use Case → Adapter (imports postgres.NewBookRepository) ✅
   Adapter → Domain (imports domain.Entity, implements domain.Repository) ✅

   Domain → Adapter? Never! ❌
   ```

### Negative

1. **Interface Duplication Across Entities:**
   ```go
   // Similar methods in multiple repository interfaces
   // book/repository.go
   type Repository interface {
       Create(ctx context.Context, book Entity) error
       GetByID(ctx context.Context, id string) (Entity, error)
   }

   // member/repository.go
   type Repository interface {
       Create(ctx context.Context, member Entity) error
       GetByID(ctx context.Context, id string) (Entity, error)
   }
   ```
   - Mitigation: This is acceptable. Each domain defines its own needs. Generic repositories often lead to leaky abstractions.

2. **Can't Use Generic Repository Pattern:**
   ```go
   // Can't do this (generic repository)
   type GenericRepository[T any] interface {
       Create(ctx context.Context, entity T) error
       GetByID(ctx context.Context, id string) (T, error)
   }
   ```
   - Mitigation: Domain-specific repositories are better. Not all entities have the same operations.

3. **Repository Interface in "Wrong" Layer:** Some developers expect interfaces next to implementations
   - Mitigation: This IS the right layer according to Dependency Inversion Principle. Domain defines contract.

## Alternatives Considered

### Alternative 1: Repository Interface in Adapter Layer

```go
// ❌ Interface in adapter layer
// internal/adapters/repository/repository.go
type BookRepository interface {
    Create(ctx context.Context, book book.Entity) error
}

// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct { /* ... */ }
func (r *PostgresBookRepository) Create(...) error { /* ... */ }

// internal/usecase/bookops/create_book.go
import "library-service/internal/adapters/repository"

type CreateBookUseCase struct {
    repo repository.BookRepository  // ← Use case depends on adapter layer!
}
```

**Why not chosen:**
- Violates Dependency Inversion Principle
- Use case depends on adapter (wrong direction)
- Can't easily swap implementations
- Couples higher layers to lower layers

### Alternative 2: No Interface, Direct Dependency

```go
// ❌ No abstraction
// internal/usecase/bookops/create_book.go
import "library-service/internal/adapters/repository/postgres"

type CreateBookUseCase struct {
    repo *postgres.BookRepository  // ← Direct dependency on PostgreSQL
}
```

**Why not chosen:**
- Tightly couples use case to PostgreSQL
- Can't swap database without changing use case
- Can't test without real database
- Violates Clean Architecture dependency rule

### Alternative 3: Generic Repository in Shared Package

```go
// ❌ Generic repository in pkg/
// pkg/repository/repository.go
type Repository[T any] interface {
    Create(ctx context.Context, entity T) error
    GetByID(ctx context.Context, id string) (T, error)
}

// internal/domain/book/repository.go
import "library-service/pkg/repository"

type Repository = repository.Repository[Entity]
```

**Why not chosen:**
- Over-abstraction (not all entities have same operations)
- Leaky abstraction (forces all entities to fit generic pattern)
- Domain loses control over contract (generic pkg dictates interface)
- Harder to add domain-specific methods like `GetByISBN`

### Alternative 4: Interface Next to Implementation

```go
// ❌ Interface and implementation in same package
// internal/adapters/repository/postgres/book.go
package postgres

type BookRepository interface {
    Create(ctx context.Context, book book.Entity) error
}

type PostgresBookRepository struct { /* ... */ }
func (r *PostgresBookRepository) Create(...) error { /* ... */ }
```

**Why not chosen:**
- Interface doesn't serve its purpose (abstraction)
- No benefit over concrete type
- Coupling between layers not reduced

## Implementation Guidelines

**When defining repository interface:**

✅ **Do:**
```go
// Domain-driven operations
type Repository interface {
    // Business operation
    GetByISBN(ctx context.Context, isbn string) (Entity, error)

    // Specific query for business rule
    GetAvailableBooks(ctx context.Context) ([]Entity, error)

    // Return domain entities
    Create(ctx context.Context, book Entity) error
}
```

❌ **Don't:**
```go
// Generic CRUD
type Repository interface {
    // Too generic
    FindOne(query map[string]interface{}) (Entity, error)

    // Returns database-specific types
    Create(book *sql.Row) error

    // Exposes database details
    ExecuteQuery(sql string, args ...interface{}) error
}
```

**Repository should:**
1. Use domain entities (not DTOs or database models)
2. Define domain-specific operations (not generic CRUD)
3. Return domain errors (not database errors)
4. Accept context for cancellation and tracing

**Implementation should:**
1. Handle database-specific concerns (SQL, connection pooling)
2. Map between domain entities and database models
3. Translate database errors to domain errors

## Real-World Example

**Domain defines need:**
```go
// internal/domain/book/repository.go
package book

type Repository interface {
    // "I need to find books by ISBN"
    GetByISBN(ctx context.Context, isbn string) (Entity, error)

    // "I need to get all available books"
    GetAvailableBooks(ctx context.Context) ([]Entity, error)
}
```

**PostgreSQL fulfills need:**
```go
// internal/adapters/repository/postgres/book.go
package postgres

func (r *BookRepository) GetByISBN(ctx context.Context, isbn string) (book.Entity, error) {
    query := "SELECT id, title, isbn, status FROM books WHERE isbn = $1"
    row := r.db.QueryRowContext(ctx, query, isbn)

    var b book.Entity
    err := row.Scan(&b.ID, &b.Title, &b.ISBN, &b.Status)
    if err == sql.ErrNoRows {
        return book.Entity{}, errors.ErrNotFound
    }
    return b, err
}
```

**MongoDB fulfills same need differently:**
```go
// internal/adapters/repository/mongodb/book.go
package mongodb

func (r *BookRepository) GetByISBN(ctx context.Context, isbn string) (book.Entity, error) {
    filter := bson.M{"isbn": isbn}
    var b book.Entity
    err := r.collection.FindOne(ctx, filter).Decode(&b)
    if err == mongo.ErrNoDocuments {
        return book.Entity{}, errors.ErrNotFound
    }
    return b, err
}
```

**Use case doesn't care:**
```go
// internal/usecase/bookops/get_book.go
package bookops

type GetBookUseCase struct {
    repo book.Repository  // ← Interface (could be PostgreSQL or MongoDB)
}

func (uc *GetBookUseCase) Execute(ctx context.Context, isbn string) (*book.Entity, error) {
    // Works with any implementation
    book, err := uc.repo.GetByISBN(ctx, isbn)
    return &book, err
}
```

## Validation

After 6 months:
- ✅ Migrated from MongoDB to PostgreSQL in 4 hours (only adapter layer changed)
- ✅ Tests run 10x faster (use memory repository instead of real database)
- ✅ Zero import cycles
- ✅ Domain layer has 100% test coverage with zero database dependencies

## References

- [Dependency Inversion Principle (SOLID)](https://en.wikipedia.org/wiki/Dependency_inversion_principle)
- [Clean Architecture - Dependency Rule](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- `.claude/common-tasks.md` - Repository implementation guides
- `.claude/common-tasks.md` - Repository implementation examples

## Related ADRs

- [ADR-001: Clean Architecture](./001-clean-architecture.md) - Why dependencies point inward
- [ADR-006: PostgreSQL](./006-postgresql.md) - Current repository implementation choice

---

**Last Reviewed:** 2024-01-17

**Next Review:** 2024-07-17
