# 🚀 AI QUICKSTART - Library Management System

> **⏱️ Time to Productivity: <2 minutes**
>
> This document is optimized for Claude Code and other AI assistants to quickly understand the codebase and become immediately productive.

## 🎯 Essential Context in 30 Seconds

**Project:** Go REST API following Clean Architecture
**Stack:** Go 1.25, PostgreSQL, Redis, Chi router, JWT auth
**Structure:** Domain → Use Case → Adapters → Infrastructure
**Key Pattern:** Use cases have "ops" suffix (e.g., `bookops` not `book`)

### 🔑 Critical Files to Know
```
internal/usecase/container.go     # Central dependency injection (ALL wiring here)
internal/adapters/http/router.go  # All HTTP routes defined here
Makefile                          # 30+ commands for everything
.claude/REFACTORING-OPPORTUNITIES.md # Recent improvements & remaining work
```

## 🎨 Common Code Patterns

### Pattern 1: Adding a New Use Case
```go
// 1. Create use case file: internal/usecase/bookops/archive_book.go
package bookops

type ArchiveBookRequest struct {
    BookID   string
    MemberID string // for authorization
}

type ArchiveBookUseCase struct {
    bookRepo    book.Repository
    bookService *book.Service
}

func (uc *ArchiveBookUseCase) Execute(ctx context.Context, req ArchiveBookRequest) error {
    logger := logutil.UseCaseLogger(ctx, "archive_book",
        zap.String("book_id", req.BookID))

    // Implementation...
    return nil
}

// 2. Wire in container.go
ArchiveBook *bookops.ArchiveBookUseCase // Add to Container struct

// 3. In NewContainer() return:
ArchiveBook: bookops.NewArchiveBookUseCase(repos.Book, bookService),
```

### Pattern 2: DTO Conversion Helper
```go
// File: internal/adapters/http/dto/book.go
func ToBookResponse(b book.Book) BookResponse {
    return BookResponse{
        ID:      b.ID,
        Name:    b.Name,
        ISBN:    b.ISBN,
        Authors: b.Authors,
    }
}

// Usage in handler
result, _ := h.useCases.GetBook.Execute(ctx, req)
response := dto.ToBookResponse(result.Book)
h.RespondJSON(w, http.StatusOK, response)
```

### Pattern 3: Request Validation
```go
func (r CreateBookRequest) Validate() error {
    if r.ISBN == "" {
        return errors.ErrValidation.WithDetails("field", "ISBN", "reason", "required")
    }
    if len(r.Name) > constants.MaxNameLength {
        return errors.ErrValidation.WithDetails("field", "name", "reason", "too long")
    }
    return nil
}

// Usage in use case
if err := req.Validate(); err != nil {
    return BookResponse{}, err
}
```

### Pattern 4: Consistent Logging
```go
// Every use case starts with:
logger := logutil.UseCaseLogger(ctx, "use_case_name",
    zap.String("key_field", req.Field))

// Log important events
logger.Info("operation succeeded", zap.String("id", result.ID))
logger.Error("operation failed", zap.Error(err))
```

## 🗺️ Decision Tree: Where to Add Code?

```
Need to add a feature?
├── Is it a new business entity?
│   ├── Yes → Start in internal/domain/{entity}/
│   └── No → Continue ↓
├── Is it a new operation on existing entity?
│   ├── Yes → Create in internal/usecase/{entity}ops/
│   └── No → Continue ↓
├── Is it a new API endpoint?
│   ├── Yes → Add to internal/adapters/http/handlers/
│   └── No → Continue ↓
├── Is it a new external integration?
│   ├── Yes → Create in internal/adapters/{service}/
│   └── No → Continue ↓
└── Is it infrastructure/config?
    └── Yes → Add to internal/infrastructure/
```

## ✅ Pre-Approved Safe Operations

These commands are safe to run without asking:
```bash
# Testing
make test              # Run all tests
make test-unit         # Unit tests only
make test-coverage     # Generate coverage report
go test ./internal/domain/...  # Test domain layer

# Code quality
make fmt               # Format code
make vet               # Run go vet
make lint              # Run golangci-lint
make ci                # Full CI pipeline

# Documentation
make gen-docs          # Regenerate Swagger docs

# Building (non-destructive)
make build             # Build all binaries
make build-api         # Build API only
```

## 🐛 Troubleshooting Checklist

### Problem: "connection refused"
```bash
make up                # Start Docker services
docker-compose -f deployments/docker/docker-compose.yml ps
```

### Problem: Tests failing
```bash
go clean -testcache    # Clear test cache
make test              # Re-run tests
```

### Problem: Import errors after adding package
```bash
go mod tidy            # Update dependencies
go mod vendor          # Update vendor directory
```

### Problem: "port already in use"
```bash
lsof -ti:8080 | xargs kill -9  # Kill process on port 8080
```

## 📊 Current State Metrics

**As of last refactoring (2025-10-09):**
- Files: 148 production, 44 test files
- Test Coverage: 100% for payment use cases
- Technical Debt: 1 TODO marker only
- Code Quality: All linting passes
- Documentation: Comprehensive in `.claude/`

## 🔄 Common Workflows

### 1. Add a Payment Feature
```bash
# 1. Create use case
touch internal/usecase/paymentops/new_feature.go

# 2. Create test
touch internal/usecase/paymentops/new_feature_test.go

# 3. Wire in container.go (see Pattern 1)

# 4. Add handler method
# Edit: internal/adapters/http/handlers/payment/manage.go

# 5. Test
make test-unit

# 6. Generate docs
make gen-docs
```

### 2. Fix a Bug
```bash
# 1. Write failing test first
go test -v -run TestSpecificCase ./path/to/package

# 2. Fix the bug

# 3. Verify fix
make test

# 4. Check no regressions
make ci
```

### 3. Add Database Migration
```bash
# 1. Create migration
make migrate-create name=add_new_column

# 2. Edit up/down SQL files in migrations/postgres/

# 3. Apply migration
make migrate-up

# 4. Test rollback
make migrate-down
make migrate-up
```

## 📁 Key Directory Purposes

```
internal/
├── domain/          # Business logic, ZERO dependencies
├── usecase/         # Orchestration, uses domain + repos
│   └── *ops/       # Operations on entities (bookops, not book)
├── adapters/        # External interfaces
│   ├── http/       # REST API handlers
│   └── repository/ # Database implementations
└── infrastructure/  # Technical plumbing (auth, server)

pkg/
├── constants/       # Centralized constants (NEW!)
├── errors/         # Custom error types
└── logutil/        # Logging utilities
```

## 🚫 Common Pitfalls to Avoid

1. **DON'T** import from outer layers in domain
2. **DON'T** put business logic in handlers
3. **DON'T** forget the "ops" suffix for use cases
4. **DON'T** use magic numbers (use pkg/constants)
5. **DON'T** skip request validation
6. **DON'T** create handlers with 8+ dependencies (use container)
7. **DON'T** manually convert DTOs in loops (use helpers)

## 💡 Quick Wins You Can Suggest

1. **Add validation** to any request struct missing Validate()
2. **Extract constants** from any magic numbers you find
3. **Create DTO helpers** for any manual conversion loops
4. **Add tests** for untested use cases
5. **Simplify complex functions** over 50 lines

## 📚 Further Reading Priority

1. `.claude/REFACTORING-OPPORTUNITIES.md` - What's been done, what's left
2. `internal/usecase/container.go` - Understand dependency wiring
3. `.claude/architecture.md` - Deep dive into Clean Architecture
4. `Makefile` - All available commands

---

**🎯 You are now ready to be productive!**

Start with: `make test` to ensure everything works, then tackle the task at hand.