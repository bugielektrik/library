# Go Project Onboarding - Quick Start Guide

> **⚡ Get productive with this Go codebase in 10 minutes**

This guide helps Claude Code instances quickly understand and work with Go projects using Clean Architecture patterns.

---

## 🚀 First 60 Seconds - Essential Context

### Project Snapshot
```bash
# Check project basics
go version                    # Go version (should be 1.21+)
go mod graph | head -5       # Key dependencies
git log --oneline -5         # Recent changes
git branch --show-current    # Current branch
```

### Quick Architecture Check
```bash
# Understand the structure (most Go projects follow similar patterns)
ls -la internal/             # Core business logic (private packages)
ls -la pkg/                  # Shared utilities (public packages)
ls -la cmd/                  # Application entry points
ls -la api/                  # API specs (OpenAPI/Swagger)
```

**Critical Questions to Answer:**
1. **Architecture Pattern:** Clean? Hexagonal? Standard Go layout?
2. **Key Entities:** What business domains exist? (Check `internal/domain/`)
3. **Entry Points:** What binaries are built? (Check `cmd/`)
4. **Testing Strategy:** Unit? Integration? (Check `*_test.go` files)

---

## 📋 10-Minute Checklist

### Step 1: Understand the Architecture (3 min)

**Read these files in order:**
1. `README.md` - Project overview
2. `.claude/architecture.md` - Architecture decisions
3. `.claude/adrs/` - Architecture Decision Records (if they exist)

**Key Patterns to Identify:**
```go
// Pattern 1: Dependency Direction
// ❌ BAD: domain imports adapters/infrastructure
import "myproject/internal/adapters/repository"

// ✅ GOOD: adapters import domain
import "myproject/internal/domain/user"

// Pattern 2: Interface Definitions
// Domain layer defines interfaces, outer layers implement them
type UserRepository interface {  // Defined in domain
    Create(ctx context.Context, user User) error
}

// Pattern 3: Use Case Structure
type CreateUserUseCase struct {
    repo UserRepository  // Depends on interface, not concrete type
}
```

### Step 2: Locate Key Components (2 min)

**Find the building blocks:**
```bash
# Domain entities (business logic)
find internal/domain -name "*.go" -type f | head -5

# Use cases (application logic)
find internal/usecase -name "*.go" -type f | head -5

# Handlers/Controllers (HTTP/gRPC)
find internal/adapters -name "*handler*.go" -o -name "*controller*.go" | head -5

# Repositories (data access)
find internal/adapters/repository -name "*.go" -type f | head -5
```

### Step 3: Run the Project (2 min)

**Startup sequence:**
```bash
# 1. Install dependencies
go mod download

# 2. Check for required services
docker-compose ps 2>/dev/null || echo "No Docker services"

# 3. Run tests to verify setup
go test ./... -short  # Skip slow tests

# 4. Start the application
go run cmd/api/main.go  # Or whatever the main entry point is
```

### Step 4: Identify Conventions (3 min)

**Naming patterns:**
```bash
# Check test file patterns
ls -1 internal/**/*_test.go | head -3

# Check mock patterns
find . -name "*mock*.go" -o -name "*fake*.go" | head -3

# Check error handling patterns
grep -r "errors\." internal/ | head -3

# Check logging patterns
grep -r "log\." internal/ | head -3
```

---

## 🎯 Common Go Project Patterns

### Pattern 1: Clean Architecture Layers

```
┌─────────────────────────────────────┐
│  Infrastructure (frameworks, DB)    │
├─────────────────────────────────────┤
│  Adapters (HTTP, gRPC, repos)       │
├─────────────────────────────────────┤
│  Use Cases (application logic)      │
├─────────────────────────────────────┤
│  Domain (entities, business rules)  │
└─────────────────────────────────────┘

Rule: Inner layers never import outer layers
```

**Directory mapping:**
- `internal/domain/` → Domain layer (core business)
- `internal/usecase/` → Use case layer (app logic)
- `internal/adapters/` → Adapter layer (HTTP, DB, cache)
- `internal/infrastructure/` → Infrastructure (frameworks, config)

### Pattern 2: Repository Pattern

```go
// Domain defines the interface
package user

type Repository interface {
    Create(ctx context.Context, user User) error
    GetByID(ctx context.Context, id string) (User, error)
}

// Adapters implement the interface
package postgres

type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
    // Implementation
}
```

### Pattern 3: Use Case Pattern

```go
package userops  // Note: "ops" suffix to avoid naming conflicts

type CreateUserRequest struct {
    Email    string
    Password string
}

type CreateUserUseCase struct {
    repo user.Repository
    hash PasswordHasher
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) error {
    // 1. Validate
    // 2. Call domain service
    // 3. Persist via repository
}
```

### Pattern 4: Dependency Injection

**Two-step wiring (common pattern):**

```go
// Step 1: Initialize infrastructure (app.go)
db := initDatabase(config)
cache := initCache(config)

// Step 2: Wire use cases (container.go)
type Container struct {
    CreateUser *CreateUserUseCase
    GetUser    *GetUserUseCase
}

func NewContainer(repos Repositories) *Container {
    return &Container{
        CreateUser: NewCreateUserUseCase(repos.User),
        GetUser:    NewGetUserUseCase(repos.User),
    }
}
```

---

## 🛠️ Essential Go Commands

### Development
```bash
# Format code
go fmt ./...
goimports -w .  # If using goimports

# Lint
golangci-lint run
go vet ./...

# Build
go build -o bin/api cmd/api/main.go

# Run
go run cmd/api/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./internal/domain/user/...

# Run with race detector
go test -race ./...

# Run only short tests
go test -short ./...
```

### Debugging
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o bin/api cmd/api/main.go

# Check dependencies
go mod tidy
go mod verify

# Find who imports a package
go mod graph | grep "package-name"

# Check for unused dependencies
go mod tidy -v
```

---

## 📝 First Task Workflow

### Adding a New Feature (Example: Add "Book" entity)

**1. Start with Domain (5 min)**
```bash
# Create domain entity
touch internal/domain/book/entity.go
touch internal/domain/book/repository.go
touch internal/domain/book/service.go
touch internal/domain/book/service_test.go
```

```go
// entity.go
package book

type Book struct {
    ID     string
    Title  string
    Author string
}

// repository.go
type Repository interface {
    Create(ctx context.Context, book Book) error
    GetByID(ctx context.Context, id string) (Book, error)
}

// service.go
type Service struct{}

func (s *Service) Validate(book Book) error {
    if book.Title == "" {
        return errors.New("title required")
    }
    return nil
}
```

**2. Add Use Case (5 min)**
```bash
# Create use case
touch internal/usecase/bookops/create_book.go
touch internal/usecase/bookops/create_book_test.go
```

```go
// create_book.go
package bookops  // Note: "ops" suffix

type CreateBookRequest struct {
    Title  string
    Author string
}

type CreateBookUseCase struct {
    repo    book.Repository
    service *book.Service
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    // 1. Create entity
    bookEntity := book.Book{
        Title:  req.Title,
        Author: req.Author,
    }

    // 2. Validate with domain service
    if err := uc.service.Validate(bookEntity); err != nil {
        return err
    }

    // 3. Persist
    return uc.repo.Create(ctx, bookEntity)
}
```

**3. Add Repository Implementation (5 min)**
```bash
touch internal/adapters/repository/postgres/book.go
```

**4. Add HTTP Handler (5 min)**
```bash
touch internal/adapters/http/handlers/book.go
```

**5. Wire Dependencies (2 min)**
```bash
# Update internal/usecase/container.go
# Update internal/infrastructure/app/app.go
```

**6. Test & Verify (3 min)**
```bash
go test ./internal/domain/book/...
go test ./internal/usecase/bookops/...
go build ./cmd/api
```

---

## 🚨 Common Pitfalls

### 1. Import Cycle Violations
```go
// ❌ NEVER: Domain imports adapters
package book
import "myproject/internal/adapters/repository"  // WRONG!

// ✅ CORRECT: Domain defines interface
package book
type Repository interface {
    Create(ctx context.Context, book Book) error
}
```

### 2. Business Logic in Wrong Layer
```go
// ❌ NEVER: Business logic in handlers
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
    if book.Price < 0 {  // Business rule in handler! WRONG!
        return
    }
}

// ✅ CORRECT: Business logic in domain service
func (s *Service) ValidateBook(book Book) error {
    if book.Price < 0 {  // Business rule in domain
        return ErrInvalidPrice
    }
    return nil
}
```

### 3. Forgetting Context
```go
// ❌ BAD: No context for cancellation/timeout
func (r *Repo) GetUser(id string) (User, error)

// ✅ GOOD: Always pass context as first parameter
func (r *Repo) GetUser(ctx context.Context, id string) (User, error)
```

### 4. Not Using Table-Driven Tests
```go
// ✅ GOOD: Table-driven tests (Go idiom)
func TestValidateBook(t *testing.T) {
    tests := []struct {
        name    string
        book    Book
        wantErr bool
    }{
        {"valid book", Book{Title: "Go"}, false},
        {"empty title", Book{Title: ""}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateBook(tt.book)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## 🔍 Project Analysis Checklist

Before starting work, answer these questions:

### Architecture
- [ ] What architecture pattern is used? (Clean/Hexagonal/Standard)
- [ ] Where do domain entities live?
- [ ] Where are use cases defined?
- [ ] How are dependencies injected?

### Database
- [ ] What database is used? (Check `docker-compose.yml` or config)
- [ ] Where are migrations? (Check `migrations/` or similar)
- [ ] How are repositories structured?

### Testing
- [ ] What testing framework? (standard library, testify, ginkgo?)
- [ ] Are mocks used? How are they generated?
- [ ] What's the coverage target? (Check CI config)

### API
- [ ] REST or gRPC? (Check `internal/adapters/`)
- [ ] Is there OpenAPI/Swagger docs? (Check `api/` folder)
- [ ] What router is used? (chi, gin, gorilla, echo?)

### Observability
- [ ] What logger? (zap, logrus, slog?)
- [ ] Are there metrics? (prometheus, statsd?)
- [ ] Is there tracing? (opentelemetry, jaeger?)

---

## 📚 Quick Reference

### File Naming Conventions
```
entity.go          # Domain entities
repository.go      # Repository interfaces
service.go         # Domain services
service_test.go    # Unit tests
*_integration_test.go  # Integration tests (with build tag)
mock_*.go          # Mock implementations
```

### Build Tags for Tests
```go
//go:build integration
// +build integration

package user_test
// Run with: go test -tags=integration
```

### Common Makefile Targets
```bash
make test          # Run tests
make lint          # Run linter
make build         # Build binaries
make run           # Run locally
make docker        # Build docker image
```

---

## ✅ You're Ready When...

You can answer these questions:

1. **Where do I add business logic?** → Domain services
2. **Where do I add API endpoints?** → Adapters (HTTP handlers)
3. **Where do I orchestrate operations?** → Use cases
4. **How do I test my code?** → Table-driven tests, mock repositories
5. **How do I run the project?** → Check README.md or Makefile

---

## 🎯 Next Steps

1. Read project-specific docs in `.claude/` folder
2. Check `CLAUDE.md` in project root for custom instructions
3. Review recent PRs to understand team conventions
4. Check CI/CD config (`.github/workflows/`) for quality gates

**Pro Tip:** When in doubt, follow the existing patterns in the codebase. Consistency > Perfection.
