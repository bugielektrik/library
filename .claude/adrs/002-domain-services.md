# ADR-002: Domain Services Pattern

**Status:** Accepted

**Date:** 2024-01-15

**Decision Makers:** Project Architecture Team

## Context

In Clean Architecture, we needed to decide where business logic should live. We had entities (structs with data), use cases (orchestration), and needed a place for complex business rules.

**Problem:** Where should logic like ISBN validation, late fee calculation, subscription eligibility, and complex business rules live?

**Constraints:**
- Must be testable without mocks (pure business logic)
- Should follow Single Responsibility Principle
- Must be reusable across multiple use cases
- Should have zero external dependencies

## Decision

We introduced **Domain Services** as stateless components containing pure business logic.

**Structure:**
```go
// internal/domain/book/service.go
type Service struct {
    // No dependencies! Pure business logic only
}

func NewService() *Service {
    return &Service{}
}

// Business rule: ISBN-13 validation with checksum
func (s *Service) ValidateISBN(isbn string) error {
    if len(isbn) != 13 {
        return errors.New("ISBN must be 13 digits")
    }

    // Checksum validation logic
    sum := 0
    for i, r := range isbn[:12] {
        digit := int(r - '0')
        if i%2 == 0 {
            sum += digit
        } else {
            sum += digit * 3
        }
    }

    checksum := (10 - (sum % 10)) % 10
    if checksum != int(isbn[12]-'0') {
        return errors.New("invalid ISBN checksum")
    }

    return nil
}

// Business rule: Calculate late fees ($0.50/day after due date)
func (s *Service) CalculateLateFee(daysLate int) float64 {
    if daysLate <= 0 {
        return 0.0
    }
    return float64(daysLate) * 0.50
}
```

**Responsibilities:**
- Domain Service: Pure business logic (validation, calculation, complex rules)
- Use Case: Orchestration (get data → validate → persist → cache)
- Entity: Data structure + simple methods

## Consequences

### Positive

1. **100% Testable Without Mocks:**
   ```go
   func TestService_CalculateLateFee(t *testing.T) {
       svc := book.NewService()

       tests := []struct {
           name     string
           daysLate int
           want     float64
       }{
           {"No late fee", 0, 0.0},
           {"One day late", 1, 0.50},
           {"Two weeks late", 14, 7.0},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               got := svc.CalculateLateFee(tt.daysLate)
               assert.Equal(t, tt.want, got)
           })
       }
   }
   ```
   No database, no mocks, no setup—just pure logic testing.

2. **Reusable Across Use Cases:**
   ```go
   // Multiple use cases can use the same business logic
   createBookUC := bookops.NewCreateBookUseCase(repo, bookService)
   updateBookUC := bookops.NewUpdateBookUseCase(repo, bookService)
   // Both use bookService.ValidateISBN()
   ```

3. **Clear Separation of Concerns:**
   - Domain Service: "WHAT is a valid ISBN?"
   - Use Case: "Get ISBN from request, validate it, save to DB, handle errors"
   - Handler: "Parse HTTP request, call use case, return HTTP response"

4. **AI-Friendly:** Clear rule: "Business logic always goes in domain service"
   - Claude Code knows exactly where to add new business rules
   - No debate about "should this go in use case or handler?"

5. **Easy to Extend:** New business rules just add methods to service
   ```go
   func (s *Service) CanBorrowBook(member Member, book Book) error {
       // Complex eligibility logic
   }
   ```

### Negative

1. **Extra Layer:** Adds one more component to understand (Entity, Service, Repository, Use Case)
   - Mitigation: Clear naming (`book.Service`) makes it obvious

2. **Potential for Anemic Domain Model:** Services can become dumping ground for all logic
   - Mitigation: Keep entity methods for simple operations (e.g., `entity.IsOverdue()`)
   - Rule: If logic needs external data, it's a service method. If it only uses entity fields, it's an entity method.

3. **When to Use Service vs Entity Method?** Can cause confusion
   - Guideline:
     ```go
     // ✅ Entity method - uses only entity fields
     func (e *LoanEntity) IsOverdue() bool {
         return time.Now().After(e.DueDate)
     }

     // ✅ Service method - needs external data or complex calculation
     func (s *Service) CalculateTotalFees(loans []LoanEntity) float64 {
         total := 0.0
         for _, loan := range loans {
             if loan.IsOverdue() {
                 days := int(time.Since(loan.DueDate).Hours() / 24)
                 total += s.CalculateLateFee(days)
             }
         }
         return total
     }
     ```

## Alternatives Considered

### Alternative 1: Rich Domain Model (all logic in entities)

**Why not chosen:**
```go
// Would require entities to have dependencies
type BookEntity struct {
    ID   string
    repo Repository  // ❌ Entity shouldn't know about persistence
}

func (e *BookEntity) Save() error {
    return e.repo.Save(e)  // ❌ Violates separation of concerns
}
```
- Entities would need dependencies (repository, cache)
- Hard to test (need to mock repository)
- Violates Single Responsibility Principle

### Alternative 2: All logic in use cases

**Why not chosen:**
```go
// CreateBookUseCase becomes huge with all validation logic
func (uc *CreateBookUseCase) Execute(req Request) error {
    // 50 lines of ISBN validation logic here
    if len(req.ISBN) != 13 { /* ... */ }

    // 30 lines of business rule validation
    if req.PublishYear > time.Now().Year() { /* ... */ }

    // Now finally create book
    book := book.NewEntity(/* ... */)
    return uc.repo.Create(ctx, book)
}
```
- Use cases become bloated (hundreds of lines)
- Business logic not reusable (duplicated across use cases)
- Hard to test business logic in isolation

### Alternative 3: Helper/Utils packages

**Why not chosen:**
```go
// utils/isbn.go
func ValidateISBN(isbn string) error { /* ... */ }

// utils/fees.go
func CalculateLateFee(days int) float64 { /* ... */ }
```
- Utils packages become dumping ground
- No clear ownership (which util function belongs to which domain?)
- Breaks domain-driven design (domain logic lives outside domain)

### Alternative 4: No services, only functions

**Why not chosen:**
```go
// domain/book/validation.go
func ValidateISBN(isbn string) error { /* ... */ }

// domain/book/calculations.go
func CalculateLateFee(days int) float64 { /* ... */ }
```
- Works for simple cases but doesn't scale
- No place to hold configuration or future dependencies
- Harder to mock in tests if needed later
- Less discoverable (scattered functions vs organized service)

## Implementation Guidelines

**When to add a method to Domain Service:**
- ✅ Validation logic (ValidateISBN, ValidateEmail)
- ✅ Calculation logic (CalculateLateFee, CalculateTotalCost)
- ✅ Business rules (CanBorrowBook, IsEligibleForSubscription)
- ✅ Complex domain operations spanning multiple entities

**When NOT to use Domain Service (use entity method instead):**
- ❌ Simple getters/setters
- ❌ Logic using only entity's own fields
- ❌ Simple state checks (IsOverdue, IsActive)

**Testing strategy:**
```go
// Domain Service tests: 100% coverage, no mocks
func TestService_ValidateISBN(t *testing.T) { /* ... */ }

// Use Case tests: Mock repository, use real domain service
func TestCreateBookUseCase(t *testing.T) {
    mockRepo := &mocks.MockRepository{}
    realService := book.NewService()  // ← Real service, no mock
    uc := bookops.NewCreateBookUseCase(mockRepo, realService)
    // Test orchestration
}
```

## Real-World Example

**Without Domain Service:**
```go
// ❌ Business logic scattered everywhere
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // Validation in handler (wrong layer!)
    if len(req.ISBN) != 13 {
        return errors.New("invalid ISBN")
    }
    // ...
}

func (uc *CreateBookUseCase) Execute(req Request) error {
    // Duplicated validation in use case
    if len(req.ISBN) != 13 {
        return errors.New("invalid ISBN")
    }
    // ...
}
```

**With Domain Service:**
```go
// ✅ Business logic centralized in domain
// domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // Single source of truth for ISBN validation
}

// usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
    // Orchestration logic
}

// usecase/bookops/update_book.go
func (uc *UpdateBookUseCase) Execute(req Request) error {
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
    // Orchestration logic
}
```

## Validation

After 6 months:
- ✅ Domain service tests have 100% coverage with zero mocks
- ✅ Business logic reused across 15+ use cases
- ✅ Adding new business rules takes 5 minutes (add method, add test)
- ✅ AI can quickly locate and modify business logic (always in `domain/{entity}/service.go`)

## References

- [Domain Services (Eric Evans - DDD)](https://www.domainlanguage.com/ddd/)
- [Anemic Domain Model vs Rich Domain Model](https://martinfowler.com/bliki/AnemicDomainModel.html)
- `.claude/common-tasks.md` - Implementation guides for domain services
- `.claude/common-tasks.md` - Examples of domain service usage

## Related ADRs

- [ADR-001: Clean Architecture](./001-clean-architecture.md) - Why we separate layers
- [ADR-003: Two-Step Dependency Injection](./003-two-step-di.md) - How services are wired

---

**Last Reviewed:** 2024-01-15

**Next Review:** 2024-07-15
