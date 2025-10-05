# Architecture Guide

> **Clean Architecture patterns and design principles**

## Overview

This project follows **Clean Architecture** (also known as Hexagonal or Onion Architecture) with strict dependency rules ensuring maintainability, testability, and flexibility.

## Core Principles

### Dependency Rule

**Dependencies point INWARD only:**

```
Domain ← Use Case ← Adapters ← Infrastructure
```

- **Domain** has ZERO external dependencies
- **Use Cases** depend only on Domain
- **Adapters** depend on Use Cases and Domain (via interfaces)
- **Infrastructure** can depend on any layer

### Layer Responsibilities

```
┌─────────────────────────────────────┐
│      Infrastructure Layer           │  External frameworks, tools
│  (Zap, PostgreSQL, Redis, Chi)     │
├─────────────────────────────────────┤
│        Adapters Layer               │  Interface implementations
│  (HTTP handlers, Repositories)      │
├─────────────────────────────────────┤
│       Use Case Layer                │  Business orchestration
│  (CreateBook, SubscribeMember)      │
├─────────────────────────────────────┤
│        Domain Layer                 │  Core business logic
│  (Entities, Services, Interfaces)   │
└─────────────────────────────────────┘
```

## Directory Structure

```
internal/
├── domain/                    # CORE: Business logic (no external deps)
│   ├── book/
│   │   ├── entity.go         # Book entity
│   │   ├── service.go        # Business rules (ISBN validation, etc.)
│   │   ├── repository.go     # Repository interface (defined here!)
│   │   ├── cache.go          # Cache interface
│   │   └── dto.go            # Domain data transfer objects
│   ├── member/
│   │   ├── entity.go
│   │   ├── service.go        # Subscription pricing logic
│   │   ├── repository.go
│   │   └── dto.go
│   └── author/
│       ├── entity.go
│       ├── repository.go
│       └── cache.go
│
├── usecase/                   # ORCHESTRATION: Application logic
│   ├── book/
│   │   ├── create_book.go    # One use case = one file
│   │   ├── update_book.go
│   │   ├── delete_book.go
│   │   ├── get_book.go
│   │   └── list_books.go
│   ├── auth/
│   │   ├── register.go
│   │   ├── login.go
│   │   └── refresh.go
│   ├── subscription/
│   │   └── subscribe_member.go
│   ├── interfaces.go          # External service interfaces
│   └── container.go           # Dependency injection container
│
├── adapters/                  # INTERFACES: External world
│   ├── http/                  # Inbound: HTTP handlers
│   │   ├── handlers/
│   │   │   ├── book.go       # Thin handlers (delegate to use cases)
│   │   │   └── auth.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── logger.go
│   │   │   └── error.go
│   │   ├── dto/              # HTTP request/response DTOs
│   │   └── router.go
│   ├── repository/            # Outbound: Data persistence
│   │   ├── postgres/         # PostgreSQL implementations
│   │   │   ├── book.go
│   │   │   ├── member.go
│   │   │   └── author.go
│   │   ├── mongo/            # MongoDB implementations
│   │   └── memory/           # In-memory (for testing)
│   ├── cache/                # Outbound: Caching
│   │   ├── redis/
│   │   └── memory/
│   ├── email/                # Outbound: Email sending
│   ├── grpc/                 # Inbound: gRPC server
│   └── storage/              # Outbound: File storage
│
└── infrastructure/            # TECHNICAL: Cross-cutting concerns
    ├── auth/
    │   ├── jwt.go            # JWT token management
    │   └── password.go       # Password hashing
    ├── store/
    │   ├── postgres.go       # Database connection
    │   └── redis.go          # Cache connection
    ├── server/
    │   └── server.go         # HTTP server setup
    ├── log/
    │   └── log.go            # Structured logging
    └── app/
        └── app.go            # Application bootstrap
```

## Key Architectural Patterns

### 1. Domain Services

**Purpose:** Encapsulate complex business logic that doesn't belong to a single entity

**Location:** `internal/domain/{entity}/service.go`

**Example - Book Service:**
```go
// internal/domain/book/service.go
type Service struct {}

func NewService() *Service {
    return &Service{}
}

// Pure business logic - no database, no HTTP
func (s *Service) ValidateISBN(isbn string) error {
    // ISBN-10 and ISBN-13 validation with checksum
    // ...
}

func (s *Service) CanBookBeDeleted(book Entity) error {
    // Business rule: can't delete if has active loans
    // ...
}
```

**Key Rules:**
- NO external dependencies (no database, no HTTP, no frameworks)
- Pure functions when possible
- 100% test coverage (easy to achieve with no deps)

### 2. Use Cases

**Purpose:** Orchestrate domain entities and services to accomplish business goals

**Location:** `internal/usecase/{entity}/{operation}.go`

**Example - Create Book Use Case:**
```go
// internal/usecase/bookops/create_book.go
type CreateBookUseCase struct {
    bookRepo    book.Repository     // Interface from domain
    bookCache   book.Cache          // Interface from domain
    bookService *book.Service       // Domain service
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*book.Entity, error) {
    // 1. Validate using domain service
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return nil, err
    }

    // 2. Create entity
    newBook := book.NewEntity(req.Name, req.ISBN, req.Genre)

    // 3. Check business rules
    if err := uc.bookService.CanCreateBook(newBook); err != nil {
        return nil, err
    }

    // 4. Persist
    if err := uc.bookRepo.Create(ctx, newBook); err != nil {
        return nil, fmt.Errorf("creating book: %w", err)
    }

    // 5. Update cache
    uc.bookCache.Set(ctx, newBook.ID, newBook)

    return &newBook, nil
}
```

**Key Rules:**
- One use case = one file (Single Responsibility Principle)
- Depends only on interfaces (defined in domain)
- Orchestrates, doesn't contain business logic
- Returns domain entities, not DTOs

### 3. Repository Pattern

**Purpose:** Abstract data persistence from business logic

**Interface:** Defined in `internal/domain/{entity}/repository.go`
**Implementation:** In `internal/adapters/repository/{type}/{entity}.go`

```go
// Domain defines the contract
// internal/domain/book/repository.go
type Repository interface {
    Create(ctx context.Context, book Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    Update(ctx context.Context, book Entity) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ListFilter) ([]Entity, error)
}

// Adapter implements it
// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct {
    db *sqlx.DB
}

func (r *PostgresBookRepository) Create(ctx context.Context, book domain.Entity) error {
    query := `INSERT INTO books (id, name, isbn, genre) VALUES ($1, $2, $3, $4)`
    _, err := r.db.ExecContext(ctx, query, book.ID, book.Name, book.ISBN, book.Genre)
    return err
}
```

**Benefits:**
- Domain is independent of database technology
- Easy to swap PostgreSQL for MongoDB (just change adapter)
- Easy to mock for testing

### 4. Dependency Injection

**Purpose:** Make dependencies explicit and testable

**Location:** `internal/usecase/container.go`

```go
// Container holds all use cases
type Container struct {
    CreateBook      *book.CreateBookUseCase
    GetBook         *book.GetBookUseCase
    RegisterMember  *auth.RegisterUseCase
    // ... other use cases
}

// Wire dependencies in constructor
func NewContainer(repos *Repositories, caches *Caches) *Container {
    bookService := book.NewService()
    memberService := member.NewService()

    return &Container{
        CreateBook: book.NewCreateBookUseCase(
            repos.Book,      // Repository interface
            caches.Book,     // Cache interface
            bookService,     // Domain service
        ),
        // ... wire other use cases
    }
}
```

**Bootstrap:** `internal/infrastructure/app/app.go`
```go
// Create all infrastructure
db := store.NewPostgres(cfg.Database)
redis := store.NewRedis(cfg.Redis)

// Create repositories
repos := &usecase.Repositories{
    Book:   postgres.NewBookRepository(db),
    Member: postgres.NewMemberRepository(db),
    Author: postgres.NewAuthorRepository(db),
}

// Create caches
caches := &usecase.Caches{
    Book:   rediscache.NewBookCache(redis),
    Author: rediscache.NewAuthorCache(redis),
}

// Wire everything
container := usecase.NewContainer(repos, caches, authSvcs)
```

### 5. Error Handling

**Domain Errors:** `pkg/errors/domain.go`
```go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrAlreadyExists = errors.New("resource already exists")
    ErrValidation    = errors.New("validation failed")
    ErrUnauthorized  = errors.New("unauthorized")
)
```

**Error Wrapping:**
```go
// GOOD: Add context and wrap
if err := repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in database: %w", err)
}

// BAD: Lose context
if err := repo.Create(ctx, book); err != nil {
    return err
}
```

## Data Flow Example

**Creating a Book:**

```
1. HTTP Request
   ↓
2. HTTP Handler (adapters/http/handlers/book.go)
   - Parse JSON to DTO
   - Call use case
   ↓
3. Use Case (usecase/bookops/create_book.go)
   - Validate using domain service
   - Create entity
   - Save via repository
   - Update cache
   ↓
4. Repository (adapters/repository/postgres/book.go)
   - Execute SQL
   - Return result
   ↓
5. Response flows back
   Repository → Use Case → Handler → HTTP Response
```

**Key Points:**
- HTTP layer never talks to database directly
- Business logic never depends on HTTP
- Each layer has clear responsibility
- Easy to test each layer in isolation

## Testing Strategy

### Unit Tests (Domain)
```go
// internal/domain/book/service_test.go
func TestService_ValidateISBN(t *testing.T) {
    svc := NewService()
    
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests (Use Cases)
```go
// internal/usecase/bookops/create_book_test.go
func TestCreateBook(t *testing.T) {
    // Use mocks
    mockRepo := mocks.NewMockRepository()
    mockCache := mocks.NewMockCache()
    svc := book.NewService()
    
    uc := NewCreateBookUseCase(mockRepo, mockCache, svc)
    
    // Test
    result, err := uc.Execute(ctx, CreateBookRequest{...})
    // Assert...
}
```

### API Tests (End-to-End)
```bash
# Use curl or HTTP test frameworks
curl -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Book","isbn":"9780132350884"}'
```

## Design Decisions

See [Architecture Decision Records](../docs/adr/) for detailed decisions:
- [ADR-001: Clean Architecture](../docs/adr/001-clean-architecture.md)
- [ADR-002: Domain Services](../docs/adr/002-domain-services.md)
- [ADR-003: Dependency Injection](../docs/adr/003-dependency-injection.md)

## Common Patterns

### Adding a New Domain

1. **Create structure:**
   ```bash
   mkdir -p internal/domain/loan
   touch internal/domain/loan/{entity.go,service.go,repository.go,dto.go}
   ```

2. **Define entity and business rules** (service.go)

3. **Define repository interface** (repository.go)

4. **Create use cases:**
   ```bash
   mkdir -p internal/usecase/loan
   touch internal/usecase/loan/{create_loan.go,return_loan.go}
   ```

5. **Implement repository:**
   ```bash
   touch internal/adapters/repository/postgres/loan.go
   ```

6. **Add HTTP handlers:**
   ```bash
   touch internal/adapters/http/handlers/loan.go
   touch internal/adapters/http/dto/loan.go
   ```

7. **Wire in container:** Update `internal/usecase/container.go`

### Changing Databases

To switch from PostgreSQL to MongoDB:
1. Implement `book.Repository` in `adapters/repository/mongo/`
2. Update `app.go` to use MongoDB repository
3. Domain and use cases remain unchanged!

## Performance Considerations

- **Caching:** Read-through cache pattern in use cases
- **Connection Pooling:** Configured in infrastructure layer
- **N+1 Queries:** Use eager loading in repositories
- **Indexes:** Defined in migrations

## References

- [The Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
