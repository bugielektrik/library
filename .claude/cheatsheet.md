# Quick Reference Card

> **Single-page reference for common patterns and commands**

## 🚀 Most Common Commands

```bash
# Daily workflow
make dev              # Start everything (docker + migrations + API)
make test             # Run all tests
make ci               # Full CI (before commit)

# Development
make run              # API only
make fmt && make lint # Fix code style
go test -v -run TestName ./path/to/package/

# Database
make migrate-up       # Apply migrations
make migrate-create name=add_something

# API Testing
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/books | jq

# Emergency fixes
lsof -ti:8080 | xargs kill -9              # Kill port 8080
make down && make up && make migrate-up    # Reset database
go clean -testcache && make test           # Clear test cache
```

## 📁 File Structure Quick Map

```
Where do I add...?

Business logic          → internal/domain/{entity}/service.go
Data validation rules   → internal/domain/{entity}/service.go
Database queries        → internal/adapters/repository/postgres/{entity}.go
API endpoint            → internal/adapters/http/handlers/{entity}.go
HTTP request/response   → internal/adapters/http/dto/{entity}.go
Use case orchestration  → internal/usecase/{entity}ops/{operation}.go
Dependency wiring       → internal/usecase/container.go
HTTP routes             → internal/adapters/http/router.go
Database migration      → make migrate-create name=xyz
Tests                   → {filename}_test.go (same directory)
```

## 🏗️ Adding New Feature (Checklist)

```bash
# 1. Domain Layer
mkdir -p internal/domain/loan
# Create: entity.go, service.go, repository.go, service_test.go

# 2. Use Case Layer
mkdir -p internal/usecase/loanops  # ← Note "ops" suffix!
# Create: create_loan.go, get_loan.go, etc.

# 3. Repository Implementation
# Create: internal/adapters/repository/postgres/loan.go

# 4. HTTP Layer
# Create: internal/adapters/http/handlers/loan.go
# Create: internal/adapters/http/dto/loan.go

# 5. Wire Dependencies
# Edit: internal/usecase/container.go
# Add to Repositories struct, Container struct, NewContainer()

# 6. Add Routes
# Edit: internal/adapters/http/router.go

# 7. Database Migration
make migrate-create name=create_loans_table

# 8. Generate API docs
make gen-docs

# 9. Test
make test && make run
```

## 💡 Code Patterns (Copy-Paste)

### Domain Entity
```go
// internal/domain/loan/entity.go
package loan

type Entity struct {
    ID        string
    BookID    string
    MemberID  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewEntity(bookID, memberID string) Entity {
    return Entity{
        ID:        uuid.New().String(),
        BookID:    bookID,
        MemberID:  memberID,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}
```

### Domain Service
```go
// internal/domain/loan/service.go
package loan

type Service struct{}

func NewService() *Service {
    return &Service{}
}

func (s *Service) ValidateDuration(d time.Duration) error {
    if d <= 0 || d > 14*24*time.Hour {
        return ErrInvalidDuration
    }
    return nil
}
```

### Repository Interface
```go
// internal/domain/loan/repository.go
package loan

import "context"

type Repository interface {
    Create(ctx context.Context, loan Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    Update(ctx context.Context, loan Entity) error
    Delete(ctx context.Context, id string) error
}
```

### Use Case
```go
// internal/usecase/loanops/create_loan.go
package loanops

import (
    "context"
    "fmt"
    "library-service/internal/domain/loan"
)

type CreateLoanUseCase struct {
    repo loan.Repository
    svc  *loan.Service
}

func NewCreateLoanUseCase(repo loan.Repository, svc *loan.Service) *CreateLoanUseCase {
    return &CreateLoanUseCase{repo: repo, svc: svc}
}

func (uc *CreateLoanUseCase) Execute(ctx context.Context, req Request) (*loan.Entity, error) {
    // Validate
    if err := uc.svc.ValidateDuration(req.Duration); err != nil {
        return nil, err
    }

    // Create entity
    newLoan := loan.NewEntity(req.BookID, req.MemberID)

    // Persist
    if err := uc.repo.Create(ctx, newLoan); err != nil {
        return nil, fmt.Errorf("creating loan: %w", err)
    }

    return &newLoan, nil
}
```

### HTTP Handler
```go
// internal/adapters/http/handlers/loan.go
package handlers

// @Summary Create loan
// @Tags loans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateLoanRequest true "Loan details"
// @Success 201 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /loans [post]
func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateLoanRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.ErrValidation, http.StatusBadRequest)
        return
    }

    loan, err := h.createLoanUC.Execute(r.Context(), req)
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }

    respondJSON(w, loan, http.StatusCreated)
}
```

### Table-Driven Test
```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "valid-input", false},
        {"invalid empty", "", true},
        {"invalid format", "bad", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🔑 Key Concepts

### Package Naming
```
Domain:   internal/domain/book      → package book
Use Case: internal/usecase/bookops  → package bookops  (ops suffix!)

WHY? Avoids import conflicts - can import both without aliases
```

### Dependency Direction
```
Domain ← Use Case ← Adapters ← Infrastructure

✅ Use case imports domain (OK)
❌ Domain imports use case (FORBIDDEN!)
```

### Error Handling
```go
// ✅ Good: Wrap with context
if err := repo.Create(ctx, entity); err != nil {
    return fmt.Errorf("creating entity: %w", err)  // %w preserves error
}

// ✅ Good: Use domain errors
return errors.ErrNotFound
return errors.ErrValidation

// ✅ Good: Check wrapped errors
if errors.Is(err, errors.ErrNotFound) {
    // handle
}
```

### Logging
```go
import "library-service/internal/infrastructure/log"

log.Info("Operation successful", "id", id, "user", user)
log.Warn("Invalid input", "field", "email")
log.Error("Failed", "error", err)
log.Debug("Processing", "item", item)
```

## 🎯 Wiring Dependencies

### Add to container.go
```go
// 1. Add repository to Repositories struct
type Repositories struct {
    // ... existing
    Loan loan.Repository  // ADD
}

// 2. Add use case to Container struct
type Container struct {
    // ... existing
    CreateLoan *loanops.CreateLoanUseCase  // ADD
}

// 3. Wire in NewContainer
func NewContainer(repos *Repositories, ...) *Container {
    loanSvc := loan.NewService()  // ADD

    return &Container{
        // ... existing
        CreateLoan: loanops.NewCreateLoanUseCase(repos.Loan, loanSvc),  // ADD
    }
}
```

### Add routes in router.go
```go
r.Route("/loans", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Post("/", handlers.Loan.CreateLoan)
    r.Get("/{id}", handlers.Loan.GetLoan)
})
```

## 🗃️ Database

### Migration Template
```sql
-- up
CREATE TABLE loans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL REFERENCES books(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_loans_book_id ON loans(book_id);

-- down
DROP TABLE IF EXISTS loans;
```

### Common Queries
```sql
-- Check migrations
SELECT * FROM schema_migrations;

-- Check tables
\dt

-- Describe table
\d books

-- Count records
SELECT COUNT(*) FROM books;
```

## 🧪 Testing

```bash
# Run specific test
go test -v -run TestCreateBook ./internal/usecase/bookops/

# With coverage
go test -coverprofile=coverage.out ./internal/domain/book/
go tool cover -html=coverage.out

# Race detection
go test -race ./...

# Integration tests
go test -tags=integration ./test/integration/

# Benchmarks
go test -bench=. -benchmem ./internal/domain/book/
```

## 🚨 Common Mistakes to Avoid

| ❌ Wrong | ✅ Right |
|---------|---------|
| `package book` in use case | `package bookops` |
| Domain imports adapter | Domain defines interface |
| Business logic in handler | Business logic in domain service |
| `return err` | `return fmt.Errorf("context: %w", err)` |
| Missing `@Security BearerAuth` | Include for protected routes |
| Skipping tests | 100% coverage for domain |
| Not running `make ci` | Always run before commit |

## 📊 Coverage Requirements

```
Domain Layer:    100% (critical business logic)
Use Cases:       80%+
Adapters:        60%+
Overall:         60%+
```

## 🔧 Troubleshooting One-Liners

```bash
# Port in use
lsof -ti:8080 | xargs kill -9

# Database won't connect
make down && make up && sleep 5

# Tests failing
go clean -testcache && make test

# Import errors
go mod tidy && goimports -w .

# Everything broken
make down && docker volume prune -f && make up && make migrate-up
```

## 📱 Swagger Essentials

```go
// Required annotations:
// @Summary      Short description
// @Tags         category
// @Accept       json
// @Produce      json
// @Security     BearerAuth        ← FOR PROTECTED ROUTES
// @Param        request body dto.Request true "Description"
// @Success      200 {object} dto.Response
// @Failure      400 {object} dto.ErrorResponse
// @Router       /path [method]

// After adding annotations:
make gen-docs
```

## 🎨 DTO Validation Tags

```go
type Request struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,max=100"`
    Age      int    `json:"age" validate:"min=0,max=150"`
    URL      string `json:"url" validate:"omitempty,url"`
    UUID     string `json:"id" validate:"required,uuid"`
}
```

## 📚 Quick Links

- Swagger UI: http://localhost:8080/swagger/index.html
- Full docs: `.claude/README.md`
- Examples: `.claude/examples/`
- Recipes: `.claude/recipes.md`
- Troubleshooting: `.claude/troubleshooting.md`

---

**Print this page and keep it visible while coding!** 📄
