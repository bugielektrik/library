# ADR 002: Introduce Domain Services for Business Logic

**Date**: 2025-10-04
**Status**: Accepted
**Decision Makers**: Development Team
**Supersedes**: Initial anemic domain model approach

## Context

The initial implementation had all business logic in use cases, resulting in:
- **Anemic domain model**: Entities were just data containers
- **Logic duplication**: Same validation logic repeated across use cases
- **Unclear ownership**: Hard to know where business rules belong
- **Poor testability**: Had to mock use cases to test business rules

Example problem:
```go
// Before: Business logic in use case (bad)
func (uc *CreateBookUseCase) Execute(ctx context.Context, input Input) error {
    // ISBN validation logic here (duplicated in UpdateBookUseCase)
    // Book validation logic here (duplicated in other use cases)
    // ...
}
```

## Decision

We will introduce **Domain Services** to encapsulate business rules that:
- Don't naturally belong to a single entity
- Are reused across multiple use cases
- Represent core business operations

### Domain Service Pattern

```go
// internal/domain/book/service.go
type Service struct {
    // No dependencies (pure business logic)
}

func NewService() *Service {
    return &Service{}
}

// Business rules as methods
func (s *Service) ValidateISBN(isbn string) error {
    // ISBN-10 and ISBN-13 validation with checksum
}

func (s *Service) ValidateBook(book Entity) error {
    // Complete book validation
}

func (s *Service) CanBookBeDeleted(book Entity) error {
    // Business rule: can't delete if borrowed
}
```

### Usage in Use Cases

```go
// Use case delegates to domain service
type CreateBookUseCase struct {
    repo        Repository
    bookService *book.Service  // Domain service
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, input Input) error {
    book := input.ToEntity()

    // Validate with domain service
    if err := uc.bookService.ValidateBook(book); err != nil {
        return err
    }

    return uc.repo.Create(ctx, book)
}
```

## Decisions

### 1. Domain Services Per Domain

Each domain package has a `service.go`:
- `internal/domain/book/service.go` - Book business rules
- `internal/domain/member/service.go` - Subscription business rules
- `internal/domain/author/service.go` - Author business rules (if needed)

### 2. No External Dependencies

Domain services are **pure Go code** with:
- ✅ No database access
- ✅ No HTTP dependencies
- ✅ No third-party libraries (except standard library)
- ✅ 100% test coverage requirement

### 3. Business Logic Checklist

A rule belongs in domain service if:
- ✅ It's a business rule (not technical)
- ✅ It's reused in multiple places
- ✅ It can be tested in isolation
- ✅ It doesn't require external data

Examples:
- ✅ ISBN validation → Domain service
- ✅ Subscription pricing → Domain service
- ❌ Database query → Repository (adapter)
- ❌ HTTP request parsing → HTTP handler (adapter)

## Implementation Examples

### Book Domain Service

```go
// ISBN validation with checksum
func (s *Service) ValidateISBN(isbn string) error {
    isbn = s.normalizeISBN(isbn)

    if len(isbn) == 10 {
        return s.validateISBN10(isbn)
    } else if len(isbn) == 13 {
        return s.validateISBN13(isbn)
    }

    return errors.New("invalid ISBN length")
}

// Business rule: can't delete borrowed books
func (s *Service) CanBookBeDeleted(book Entity) error {
    if len(book.Members) > 0 {
        return errors.New("cannot delete book: currently borrowed")
    }
    return nil
}
```

### Member Domain Service

```go
// Subscription pricing with bulk discounts
func (s *Service) CalculateSubscriptionPrice(subType string, months int) (float64, error) {
    basePrice := s.getMonthlyPrice(subType)
    total := basePrice * float64(months)

    // Business rules: bulk discounts
    if months >= 12 {
        total *= 0.8  // 20% discount for annual
    } else if months >= 6 {
        total *= 0.9  // 10% discount for 6+ months
    }

    return total, nil
}

// Business rule: no downgrades
func (s *Service) CanUpgradeSubscription(current, target string) error {
    if current == "premium" && target == "basic" {
        return errors.New("cannot downgrade from premium to basic")
    }
    return nil
}
```

## Consequences

### Positive

✅ **Single Responsibility**: Business logic centralized in domain services
✅ **Testability**: Can test business rules without mocking anything
✅ **Reusability**: Same logic used across all use cases
✅ **Clarity**: Clear where business rules live
✅ **Vibecoding**: Claude Code can easily find and modify business logic
✅ **Coverage**: Easy to achieve 100% test coverage on critical logic

### Negative

❌ **More Files**: Additional service files per domain
❌ **Learning Curve**: Team needs to distinguish entity vs service logic
❌ **Potential Overuse**: Risk of putting everything in services

### Mitigations

- Clear guidelines on what belongs in domain service
- Code review to prevent service bloat
- Examples in documentation
- Regular refactoring to move logic from use cases to services

## Testing Strategy

Domain services require **100% test coverage**:

```go
func TestService_ValidateISBN(t *testing.T) {
    service := NewService()

    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-10", "0-306-40615-2", false},
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-0", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateISBN() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Migration Path

### Before (Anemic Domain)

```go
// Use case does everything
func (uc *CreateBookUseCase) Execute(ctx, input) error {
    // Validation logic here (duplicated)
    if input.Name == "" {
        return errors.New("name required")
    }

    // ISBN validation here (duplicated)
    isbn := cleanISBN(input.ISBN)
    if !isValidISBN(isbn) {
        return errors.New("invalid ISBN")
    }

    book := Book{...}
    return uc.repo.Create(ctx, book)
}
```

### After (Rich Domain with Services)

```go
// Use case delegates to domain service
func (uc *CreateBookUseCase) Execute(ctx, input) error {
    book := input.ToEntity()

    // Domain service handles all validation
    if err := uc.bookService.ValidateBook(book); err != nil {
        return err
    }

    return uc.repo.Create(ctx, book)
}

// Domain service (reusable, testable)
func (s *Service) ValidateBook(book Entity) error {
    if err := s.ValidateISBN(book.ISBN); err != nil {
        return err
    }

    if book.Name == "" {
        return errors.New("name required")
    }

    if len(book.Authors) == 0 {
        return errors.New("at least one author required")
    }

    return nil
}
```

## Alternatives Considered

### 1. Keep Logic in Use Cases (Anemic Domain)
- ❌ Duplication across use cases
- ❌ Hard to find business rules
- ❌ Lower testability

### 2. Put Logic in Entities (Rich Entities)
- ✅ Encapsulation
- ❌ Entities become bloated
- ❌ Hard to test complex logic involving multiple entities

### 3. Separate Validation Package
- ✅ Centralized validation
- ❌ Separated from domain
- ❌ Not domain-driven

## References

- [Domain Services (DDD)](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Anemic Domain Model](https://martinfowler.com/bliki/AnemicDomainModel.html)
- [Rich vs Anemic Domain Models](https://www.baeldung.com/java-anemic-vs-rich-domain-objects)

## Review Notes

**Impact**: All existing use cases refactored to use domain services

**Benefits Realized**:
- Book domain: ISBN validation, book safety checks
- Member domain: Subscription pricing, upgrade rules, grace periods

**Next Steps**:
- Monitor for service bloat
- Review entity vs service boundary every 3 months
- Add more domain services as new business rules emerge

**Next Review**: After 3 months
