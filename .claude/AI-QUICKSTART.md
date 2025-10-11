# Quick Start for AI Assistants

**🎯 Goal:** Get productive in 60 seconds, expert in 5 minutes

**Last Updated:** 2025-10-09

---

## ⚡ First 60 Seconds (CRITICAL - READ THIS)

### 1. Verify Project Status

```bash
make test && make build
```

**If this fails, STOP and read:** `.claude/troubleshooting.md`

### 2. Understand Architecture

```
Domain → Use Case → Adapters → Infrastructure
  ↓         ↓           ↓            ↓
Pure     Orchestr   HTTP/DB      Config
Logic     ation                  Logging
```

**Key Rule:** Domain NEVER imports from outer layers

### 3. Key Files (Read These First)

| File | Lines | Purpose | When to Read |
|------|-------|---------|--------------|
| `internal/usecase/container.go` | 1-197 | **Complete DI guide + architecture** | Adding ANY feature |
| `cmd/api/main.go` | 28-120 | Boot sequence documentation | Understanding startup |
| `.claude/context-guide.md` | All | Task-specific reading lists | Before starting task |

---

## 📚 Documentation by Task

### Adding Features

| Task | Read This |
|------|-----------|
| New domain entity | [.claude/examples/adding-domain-entity.md](./examples/adding-domain-entity.md) |
| New API endpoint | [.claude/examples/adding-api-endpoint.md](./examples/adding-api-endpoint.md) |
| Integration tests | [.claude/examples/integration-testing.md](./examples/integration-testing.md) |

### Understanding Code

| Component | Read This |
|-----------|-----------|
| Book domain | `internal/domain/book/doc.go` |
| Payment flow | `internal/domain/payment/doc.go` |
| Auth system | `internal/adapters/http/handlers/auth/doc.go` |
| Database | `internal/adapters/repository/postgres/doc.go` |

### Problem Solving

| Issue | Read This |
|-------|-----------|
| Build failures | [.claude/troubleshooting.md](./troubleshooting.md) |
| Test failures | [.claude/testing.md](./testing.md) |
| Common mistakes | [.claude/gotchas.md](./gotchas.md) |
| Quick answers | [.claude/examples/common-tasks.md](./examples/common-tasks.md) |

---

## 🎨 Copy-Paste Templates

**Never start from scratch - copy existing patterns!**

| Need | Copy From | Key Pattern |
|------|-----------|-------------|
| **Use Case** | `internal/usecase/bookops/create_book.go` | repo + cache + domain service |
| **HTTP Handler** | `internal/adapters/http/handlers/book/crud.go` | decode → validate → execute → respond |
| **Domain Service** | `internal/domain/book/service.go` | Pure functions, zero dependencies |
| **DTO** | `internal/adapters/http/dto/book.go` | Request + Response + conversions |
| **Use Case Test** | `internal/usecase/bookops/create_book_test.go` | Table-driven + gomock |
| **Repository** | `internal/adapters/repository/postgres/book.go` | BaseRepository[T] pattern |
| **Domain Entity** | `internal/domain/book/book.go` | Plain struct + business rules |

---

## 🔍 Code Navigation Cheatsheet

### Finding Things Fast

```bash
# Find use case by name
find internal/usecase -name "*create_book*"

# Find handler for endpoint
grep -r "POST /books" internal/adapters/http/handlers

# Find domain service
ls internal/domain/book/service.go

# Find all tests for a feature
find internal -name "*book*test.go"

# Find DTO definitions
ls internal/adapters/http/dto/
```

### Package Structure

```
internal/
├── domain/              # Business logic (pure, no deps)
│   ├── book/           # Entity, service, repository interface
│   ├── member/
│   ├── payment/
│   └── reservation/
│
├── usecase/            # Orchestration (depends on domain)
│   ├── bookops/        # Note: "ops" suffix avoids naming conflicts
│   ├── authops/
│   ├── paymentops/
│   └── reservationops/
│
├── adapters/           # External interfaces
│   ├── http/           # HTTP handlers, DTOs, middleware
│   ├── repository/     # Postgres, memory, mocks
│   └── cache/          # Redis, memory
│
└── infrastructure/     # Technical concerns
    ├── app/            # Application bootstrap
    ├── auth/           # JWT + password services
    ├── config/         # Configuration loading
    └── server/         # HTTP server setup
```

---

## 🚀 Common Tasks - Quick Commands

### Development

```bash
# Full development stack
make dev                # Docker + migrations + API server

# Individual operations
make build              # Build all binaries
make run                # Run API server
make test               # Run all tests
make ci                 # Full CI pipeline locally
```

### Testing

```bash
# Quick tests (no DB)
make test-unit          # or: go test ./... -short

# Integration tests (requires DB)
make test-integration

# Specific package
go test ./internal/usecase/bookops/... -v

# Specific test
go test ./internal/domain/book/... -run TestValidateISBN

# With coverage
make test-coverage      # Opens HTML report
```

### Database

```bash
# Migrations
make migrate-up         # Apply all pending
make migrate-down       # Rollback last one
make migrate-create name=add_ratings  # Create new migration

# Database access
psql -h localhost -U library -d library
# Password: library123
```

### Code Quality

```bash
# Before committing (ALWAYS RUN THIS)
make ci                 # fmt → vet → lint → test → build

# Individual checks
make fmt                # Format code
make lint               # Run linter
make vet                # Static analysis
```

### API Documentation

```bash
# Regenerate Swagger docs
make gen-docs

# View docs
# Start server, then visit: http://localhost:8080/swagger/index.html
```

---

## 🎯 Architecture Patterns

### Use Case Pattern (Standard CRUD)

```go
// File: internal/usecase/bookops/create_book.go

type CreateBookUseCase struct {
	repo    book.Repository      // Database
	cache   book.Cache           // Caching
	service *book.Service        // Business rules
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*CreateBookResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "create_book", zap.String("name", req.Name))

	// 1. Validate with domain service
	if err := uc.service.ValidateISBN(req.ISBN); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return nil, errors.ErrValidation.WithMessage("invalid ISBN")
	}

	// 2. Create domain entity
	bookEntity := book.New(req.Name, req.ISBN, req.Authors)

	// 3. Persist to repository
	id, err := uc.repo.Create(ctx, bookEntity)
	if err != nil {
		logger.Error("failed to create book", zap.Error(err))
		return nil, fmt.Errorf("create book: %w", err)
	}

	// 4. Update cache
	_ = uc.cache.Set(ctx, id, bookEntity)  // Cache errors are non-fatal

	logger.Info("book created", zap.String("id", id))
	return &CreateBookResponse{Book: bookEntity}, nil
}
```

### HTTP Handler Pattern

```go
// File: internal/adapters/http/handlers/book/crud.go

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "create")

	// 1. Decode request
	var req dto.CreateBookRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// 2. Validate
	if !h.validator.ValidateStruct(w, req) {
		return  // Validator already sent error response
	}

	// 3. Execute use case
	result, err := h.useCases.CreateBook.Execute(ctx, bookops.CreateBookRequest{
		Name:    req.Name,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// 4. Convert to DTO and respond
	response := dto.FromBookResponse(result.Book)
	logger.Info("book created via API", zap.String("id", response.ID))
	h.RespondJSON(w, http.StatusCreated, response)
}
```

### Domain Service Pattern

```go
// File: internal/domain/book/service.go

// Service provides business logic for books.
// ZERO external dependencies - pure functions only.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// ValidateISBN checks if ISBN is valid according to ISBN-13 standard.
func (s *Service) ValidateISBN(isbn string) error {
	// Pure business logic - no database, no HTTP, no frameworks
	if len(isbn) != 17 {  // Format: 978-0-306-40615-7
		return errors.New("ISBN must be 17 characters")
	}
	// ... validation logic
	return nil
}
```

---

## ⚠️ Common Mistakes (AVOID THESE!)

### ❌ Domain Importing from Outer Layers

```go
// BAD - Domain importing from use case
package book

import "library-service/internal/usecase/bookops"  // ❌ WRONG!

// GOOD - Domain has ZERO imports from outer layers
package book

import "context"  // ✅ OK - standard library only
```

### ❌ Creating Domain Services in app.go

```go
// BAD
// File: internal/infrastructure/app/app.go
bookService := book.NewService()  // ❌ WRONG! Too early

// GOOD
// File: internal/usecase/container.go
func NewContainer(...) *Container {
    bookService := book.NewService()  // ✅ Correct place
}
```

### ❌ Using "v1" as Package Name

```go
// BAD
package v1  // ❌ Not descriptive

// GOOD
package book     // ✅ Domain-specific
package auth     // ✅ Clear purpose
package payment  // ✅ Meaningful name
```

### ❌ Testing Use Cases Without Mocks

```go
// BAD - Using real repository in use case test
bookRepo := postgres.NewBookRepository(db)  // ❌ Requires database

// GOOD - Using mock
ctrl := gomock.NewController(t)
mockRepo := mocks.NewMockBookRepository(ctrl)  // ✅ No dependencies
mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("id", nil)
```

### ❌ Committing Without Tests

```go
// BAD - New use case without tests
git commit -m "Add CreatePayment use case"  // ❌ No tests!

// GOOD - Tests included
// Files: create_payment.go + create_payment_test.go
git commit -m "Add CreatePayment use case with tests"  // ✅ Tested
```

---

## 🔧 Adding a New Feature (Quick Workflow)

### Example: Adding a "Rating" feature

**1. Domain Layer (30 min)**
```bash
# Create files
internal/domain/rating/
├── rating.go           # Entity
├── service.go          # Business logic
├── repository.go       # Interface
└── doc.go              # Package docs
```

**2. Use Case Layer (45 min)**
```bash
# Create files
internal/usecase/ratingops/
├── create_rating.go
├── get_rating.go
├── list_ratings.go
└── doc.go

# Wire in container.go:
# - Add to Repositories struct
# - Add to Container struct
# - Create service in NewContainer()
# - Wire use cases
```

**3. Adapter Layer (60 min)**
```bash
# Repository
internal/adapters/repository/postgres/rating.go

# HTTP Handler
internal/adapters/http/handlers/rating/
├── handler.go
├── crud.go
└── doc.go

# DTOs
internal/adapters/http/dto/rating.go
```

**4. Migration (10 min)**
```bash
make migrate-create name=create_ratings_table
# Edit migrations/postgres/XXXXXX_create_ratings_table.up.sql
make migrate-up
```

**5. Wire Everything (10 min)**
```bash
# 1. internal/usecase/container.go (add repository + use cases)
# 2. internal/infrastructure/app/app.go (create repository)
# 3. internal/adapters/http/router.go (register handler)
```

**6. Test & Document (30 min)**
```bash
# Add tests
internal/usecase/ratingops/create_rating_test.go

# Update Swagger
make gen-docs

# Verify
make ci
```

**Total:** ~3 hours for complete feature

---

## 📊 Project Statistics

```
Go Files:      202 (175 production, 27 tests)
Directories:   47
Largest Files: dto/payment.go (507 lines)
               container.go (406 lines)
Test Coverage: ~85% (domain + tested use cases)
Build Time:    < 5 seconds
Test Time:     < 5 seconds (unit tests)
```

---

## 🎓 Learning Path

### Beginner (New to this codebase)

1. Read this file (you're doing it!)
2. Run `make ci` to verify setup
3. Read `.claude/examples/adding-api-endpoint.md`
4. Follow the example to add a simple endpoint
5. Read `.claude/architecture.md` for deeper understanding

### Intermediate (Comfortable with basics)

1. Read `.claude/examples/adding-domain-entity.md`
2. Add a complete domain entity with tests
3. Read `internal/usecase/container.go` (lines 1-197)
4. Understand dependency injection patterns

### Advanced (Contributing complex features)

1. Read all `.claude/examples/*.md` files
2. Study payment domain (most complex)
3. Read `.claude/debugging-guide.md`
4. Read `.claude/development-workflows.md`

---

## 🔗 Quick Links

### Documentation
- **Main docs:** `.claude/README.md`
- **Architecture:** `.claude/architecture.md`
- **Standards:** `.claude/standards.md`
- **Testing:** `.claude/testing.md`
- **API:** `.claude/api.md`

### Examples
- **Domain Entity:** `.claude/examples/adding-domain-entity.md`
- **API Endpoint:** `.claude/examples/adding-api-endpoint.md`
- **Integration Tests:** `.claude/examples/integration-testing.md`
- **Common Tasks:** `.claude/examples/common-tasks.md`

### Reference
- **Commands:** `.claude/commands.md`
- **Cheatsheet:** `.claude/cheatsheet.md`
- **FAQ:** `.claude/faq.md`
- **Troubleshooting:** `.claude/troubleshooting.md`

---

## ✅ Pre-Commit Checklist

Before committing, verify:

- [ ] `make ci` passes (fmt, lint, test, build)
- [ ] Added tests for new use cases
- [ ] Updated Swagger if API changed (`make gen-docs`)
- [ ] Added logging to use cases/handlers
- [ ] Followed existing patterns (copied from templates)
- [ ] No domain layer imports from outer layers
- [ ] Package names are domain-specific (not "v1")

---

## 💡 Pro Tips

1. **Always run `make ci` before committing** - Catches 90% of issues
2. **Copy existing code** - Don't reinvent patterns
3. **Read container.go first** - It explains everything about DI
4. **Use table-driven tests** - Standard Go testing pattern
5. **Mock repositories** - Use case tests should be fast
6. **Real domain services** - They're pure functions, no mocking needed
7. **Check examples/** - Complete working code for common tasks
8. **Domain has no deps** - If you're importing from outer layers, you're wrong

---

## 🆘 When Stuck

1. **Check examples:** `.claude/examples/common-tasks.md` (covers 90% of questions)
2. **Read troubleshooting:** `.claude/troubleshooting.md`
3. **Check gotchas:** `.claude/gotchas.md`
4. **Review similar code:** Find existing example of what you're trying to do
5. **Verify environment:** `make test && make build`

---

## 🎯 Success Metrics

**You're productive when you can:**
- Add a new API endpoint in < 1 hour ✅
- Understand any use case in < 5 minutes ✅
- Find relevant code in < 60 seconds ✅
- Write tests without looking at docs ✅
- Navigate codebase confidently ✅

**This guide should get you there. Happy coding! 🚀**

---

**Last Updated:** 2025-10-09
**Maintained By:** Development Team
**Questions?** Check `.claude/faq.md` or `.claude/troubleshooting.md`
