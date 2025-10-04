# Domain Layer

**Core business logic and entities - the heart of Clean Architecture.**

## Purpose

This layer contains:
- **Entities**: Core business objects (Book, Member, Author)
- **Domain Services**: Complex business rules that don't belong to a single entity
- **Interfaces**: Repository and cache contracts (implemented by outer layers)
- **Value Objects**: Immutable domain concepts
- **Business Rules**: ISBN validation, subscription pricing, business constraints

## Dependency Rule

**The domain layer has ZERO dependencies on outer layers.**

```
Domain (this layer)
  ↑ depends on
  ✗ NO dependencies (pure business logic)
```

## Directory Structure

```
domain/
├── book/              # Book domain
│   ├── entity.go      # Book entity
│   ├── service.go     # ISBN validation, business rules
│   ├── repository.go  # Repository interface
│   ├── cache.go       # Cache interface
│   └── dto.go         # Domain DTOs
│
├── member/            # Member domain
│   ├── entity.go      # Member entity
│   ├── service.go     # Subscription pricing, validation
│   ├── repository.go  # Repository interface
│   └── dto.go         # Domain DTOs
│
└── author/            # Author domain
    ├── entity.go      # Author entity
    ├── repository.go  # Repository interface
    ├── cache.go       # Cache interface
    └── dto.go         # Domain DTOs
```

## Domain Services

### Book Service (`book/service.go`)

Business rules for book management:
- `ValidateISBN(isbn string) error` - Validates ISBN-10/ISBN-13 with checksum
- `ValidateBook(book Entity) error` - Complete book validation
- `CanBookBeDeleted(book Entity) error` - Deletion safety check
- `NormalizeISBN(isbn string) (string, error)` - ISBN-10 → ISBN-13 conversion

```go
service := book.NewService()
if err := service.ValidateISBN("978-0-306-40615-7"); err != nil {
    return err
}
```

### Member Service (`member/service.go`)

Subscription and membership business logic:
- `CalculateSubscriptionPrice(type string, months int) (float64, error)` - Pricing with bulk discounts
- `ValidateSubscriptionType(type string) error` - Type validation
- `CanUpgradeSubscription(current, target string) error` - Upgrade rules (no downgrades)
- `IsWithinGracePeriod(expiresAt time.Time, type string) bool` - Grace period check

```go
service := member.NewService()
price, _ := service.CalculateSubscriptionPrice("premium", 12) // 20% bulk discount
```

## Repository Pattern

Repositories are **interfaces defined in domain, implemented in adapters**.

```go
// Domain defines interface
type Repository interface {
    Create(ctx context.Context, book Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    // ... other methods
}

// Adapters implement it
// internal/adapters/repository/book_postgres.go
type PostgresBookRepository struct { ... }

func (r *PostgresBookRepository) Create(ctx context.Context, book domain.Entity) error {
    // Implementation
}
```

## Testing

Domain layer should have **100% test coverage** as it contains critical business logic.

```bash
# Run domain tests
go test ./internal/domain/... -cover

# Benchmark domain services
go test -bench=. ./internal/domain/book
```

## Design Principles

1. **Business-First**: Domain models business, not database structure
2. **Framework-Independent**: No Gin, no GORM, just pure Go
3. **Testable**: No external dependencies, easy to test
4. **Rich Domain Model**: Logic lives here, not scattered in use cases
5. **Immutability**: Value objects are immutable
6. **Interface Segregation**: Small, focused interfaces

## Adding New Domain

To add a new domain (e.g., `loan`):

1. **Create directory**: `internal/domain/loan/`
2. **Define entity**: `entity.go`
3. **Add business logic**: `service.go` (if needed)
4. **Define interfaces**: `repository.go`, `cache.go`
5. **Write tests**: `service_test.go` (100% coverage)
6. **Add package doc**: `doc.go`

## References

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Package Documentation](./doc.go)
