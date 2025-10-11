# ADR 003: Domain Services vs Infrastructure Services

**Status:** Accepted

**Date:** 2025-10-09

**Context:**

In Clean Architecture, "services" appear in two different layers:

1. **Domain Services** - Business logic that doesn't naturally belong to a single entity
2. **Infrastructure Services** - Technical concerns like authentication, password hashing

This creates confusion about:
- Where to create each type of service
- When to pass services as dependencies vs create them
- Why some services are in `app.go` and others in `container.go`

**The Problem:**

```go
// Where does this belong?
type BookService struct {
    db *sqlx.DB  // Database dependency
}

func (s *BookService) ValidateISBN(isbn string) error {
    // Check if ISBN already exists in database
    var count int
    s.db.QueryRow("SELECT COUNT(*) FROM books WHERE isbn = $1", isbn).Scan(&count)
    if count > 0 {
        return errors.New("ISBN already exists")
    }
    return nil
}
```

**Why this is wrong:**
- Business logic (ISBN validation) mixed with infrastructure (database)
- Can't test without database
- Domain layer depends on external library

## Decision

### Domain Services (Business Layer)

**Definition:** Encapsulate business logic that involves multiple entities or doesn't naturally belong to one entity

**Characteristics:**
- **Pure functions:** No external dependencies
- **Stateless:** No database, HTTP, or framework dependencies
- **Created in:** `internal/usecase/container.go` `NewContainer()` function
- **Location:** `internal/domain/{entity}/service.go`

**Examples:**
- ISBN validation (format rules)
- Subscription pricing calculation
- Payment state machine transitions
- Reservation expiration logic

```go
// internal/domain/book/service.go
package book

import "errors"

// Service provides business logic for books.
type Service struct{}

func NewService() *Service {
    return &Service{}
}

// ValidateISBN validates ISBN format according to ISBN-13 standard.
// Pure business logic - no database, no HTTP, no frameworks.
func (s *Service) ValidateISBN(isbn string) error {
    // Business rule: ISBN-13 format is 978-X-XXX-XXXXX-X
    if len(isbn) != 17 {
        return errors.New("ISBN must be 17 characters")
    }

    if !strings.HasPrefix(isbn, "978-") && !strings.HasPrefix(isbn, "979-") {
        return errors.New("ISBN must start with 978- or 979-")
    }

    // Calculate checksum (business logic)
    // ...
    return nil
}
```

### Infrastructure Services (Technical Layer)

**Definition:** Handle technical concerns like authentication, encryption, external API clients

**Characteristics:**
- **External dependencies:** JWT libraries, bcrypt, HTTP clients
- **Configuration-driven:** Secrets, URLs, timeouts
- **Created in:** `internal/infrastructure/app/app.go` (application bootstrap)
- **Location:** `internal/infrastructure/{concern}/`

**Examples:**
- JWT token generation/validation
- Password hashing/verification
- Payment gateway client
- Email sending service

```go
// internal/infrastructure/auth/jwt.go
package auth

import (
    "github.com/golang-jwt/jwt/v5"  // External library OK
    "time"
)

type JWTService struct {
    secretKey []byte       // Configuration
    expiry    time.Duration
}

func NewJWTService(secretKey string, expiry time.Duration) *JWTService {
    return &JWTService{
        secretKey: []byte(secretKey),
        expiry:    expiry,
    }
}

func (s *JWTService) GenerateToken(memberID string) (string, error) {
    // Technical implementation using external library
    claims := jwt.MapClaims{
        "member_id": memberID,
        "exp":       time.Now().Add(s.expiry).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}
```

## Implementation Pattern

### Two-Step Wiring

**Step 1: app.go - Create Infrastructure Services**

```go
// internal/infrastructure/app/app.go
package app

func NewApp(config *Config) *App {
    // 1. Database connections (infrastructure)
    db := connectPostgreSQL(config.DatabaseURL)

    // 2. Repository implementations (adapters)
    bookRepo := postgres.NewBookRepository(db)
    memberRepo := postgres.NewMemberRepository(db)

    // 3. Infrastructure services (technical concerns)
    jwtService := auth.NewJWTService(config.JWTSecret, config.JWTExpiry)
    passwordService := auth.NewPasswordService()
    paymentGateway := epayment.NewGateway(config.PaymentConfig)

    authServices := &usecase.AuthServices{
        JWTService:      jwtService,
        PasswordService: passwordService,
    }

    gatewayServices := &usecase.GatewayServices{
        PaymentGateway: paymentGateway,
    }

    // 4. Pass to container (domain services created there)
    container := usecase.NewContainer(repos, caches, authServices, gatewayServices)

    return &App{Container: container}
}
```

**Step 2: container.go - Create Domain Services**

```go
// internal/usecase/container.go
package usecase

func NewContainer(repos *Repositories, caches *Caches, authSvcs *AuthServices, gatewaySvcs *GatewayServices) *Container {
    // Create domain services (pure business logic, no config needed)
    bookService := book.NewService()
    memberService := member.NewService()
    paymentService := payment.NewService()

    return &Container{
        // Use cases get both types of services
        CreateBook: bookops.NewCreateBookUseCase(
            repos.Book,        // Repository
            bookService,       // Domain service (created here)
        ),
        RegisterMember: authops.NewRegisterUseCase(
            repos.Member,           // Repository
            authSvcs.PasswordService, // Infrastructure service (passed in)
            authSvcs.JWTService,      // Infrastructure service (passed in)
            memberService,          // Domain service (created here)
        ),
    }
}
```

## Decision Matrix

| Characteristic | Domain Service | Infrastructure Service |
|----------------|----------------|------------------------|
| **Dependencies** | None (pure) | External libraries |
| **Created in** | `container.go` | `app.go` |
| **Location** | `internal/domain/` | `internal/infrastructure/` |
| **Configuration** | No config needed | Requires secrets, URLs |
| **Testability** | Test directly | Mock in use case tests |
| **Examples** | ISBN validation, pricing | JWT, bcrypt, HTTP clients |

## Consequences

### Positive

✅ **Clear separation:** Business vs technical concerns

✅ **Easy testing:**
```go
// Domain service - test directly
func TestValidateISBN(t *testing.T) {
    service := book.NewService()  // No dependencies!
    err := service.ValidateISBN("978-0-306-40615-7")
    assert.NoError(t, err)
}

// Infrastructure service - mock in tests
func TestRegister(t *testing.T) {
    mockJWT := mocks.NewMockJWTService()
    uc := authops.NewRegisterUseCase(repo, passwordSvc, mockJWT, memberSvc)
}
```

✅ **Domain stays pure:** No framework coupling

✅ **Infrastructure reusable:** JWT service used across multiple use cases

### Negative

❌ **Two creation locations:** Must remember where to create each type

❌ **More parameters:** Infrastructure services passed through multiple layers

## Common Mistakes

### ❌ MISTAKE 1: Creating domain services in app.go

```go
// internal/infrastructure/app/app.go - WRONG!
bookService := book.NewService()  // Domain service created too early
```

**Why wrong:** Domain services are lightweight and don't need early initialization

**Fix:** Create in `container.go`

### ❌ MISTAKE 2: Adding dependencies to domain services

```go
// internal/domain/book/service.go - WRONG!
type Service struct {
    db *sqlx.DB  // External dependency!
}
```

**Why wrong:** Domain service should be pure business logic

**Fix:** Move database query to repository, keep service pure

### ❌ MISTAKE 3: Creating infrastructure services in container.go

```go
// internal/usecase/container.go - WRONG!
jwtService := auth.NewJWTService(config.JWTSecret, config.JWTExpiry)  // Config not available here!
```

**Why wrong:** Configuration should be loaded once at app startup

**Fix:** Create in `app.go`, pass to container

## Examples

### Example 1: Payment State Transitions

**Domain Service:**
```go
// internal/domain/payment/service.go
func (s *Service) CanTransitionTo(current, target Status) bool {
    // Business rules for payment state machine
    transitions := map[Status][]Status{
        StatusPending:    {StatusProcessing, StatusCancelled, StatusFailed},
        StatusProcessing: {StatusCompleted, StatusFailed},
        StatusCompleted:  {StatusRefunded},
    }
    allowed := transitions[current]
    for _, status := range allowed {
        if status == target {
            return true
        }
    }
    return false
}
```

**Why domain service:** Pure business rule, no external dependencies

### Example 2: Password Hashing

**Infrastructure Service:**
```go
// internal/infrastructure/auth/password.go
func (s *PasswordService) HashPassword(password string) (string, error) {
    // Uses bcrypt external library
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}
```

**Why infrastructure service:** Depends on external library (bcrypt)

## Related Decisions

- **ADR 002:** Clean Architecture Boundaries - Layer dependency rules
- **ADR 001:** Use Case "ops" Suffix - How use cases are organized

## References

- **Implementation:** `internal/usecase/container.go` (comprehensive documentation)
- **Example:** Payment domain service (`internal/domain/payment/service.go`)
- **Example:** JWT infrastructure service (`internal/infrastructure/auth/jwt.go`)

## Notes for AI Assistants

### When creating a new service, ask:

**Does it have external dependencies?**
- **Yes** (database, HTTP, JWT, bcrypt) → Infrastructure service in `app.go`
- **No** (pure logic) → Domain service in `container.go`

**Does it need configuration?**
- **Yes** (secrets, URLs, timeouts) → Infrastructure service in `app.go`
- **No** → Domain service in `container.go`

### Quick Reference

```go
// DOMAIN SERVICE - Pure business logic
type BookService struct{}  // No fields!

func (s *BookService) ValidateISBN(isbn string) error {
    // Pure validation logic
}

// INFRASTRUCTURE SERVICE - External dependencies
type JWTService struct {
    secret []byte        // Configuration
    client *http.Client  // External dependency
}

func (s *JWTService) GenerateToken(userID string) (string, error) {
    // Uses external JWT library
}
```

## Revision History

- **2025-10-09:** Initial ADR documenting service creation patterns
