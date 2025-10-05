# Frequently Asked Questions (FAQ)

> **Quick answers to common questions**

## 🏗️ Architecture Questions

### Q: Why use "ops" suffix for use case packages?

**A:** To avoid naming conflicts with domain packages.

```go
// Without suffix - conflict!
import (
    "library-service/internal/domain/book"      // package book
    "library-service/internal/usecase/book"     // package book ← CONFLICT!
)

// With "ops" suffix - clean!
import (
    "library-service/internal/domain/book"      // package book
    "library-service/internal/usecase/bookops"  // package bookops ← Different!
)

// No need for import aliases
book.Entity{}           // From domain
bookops.CreateBookUseCase{}  // From use case
```

### Q: Where does business logic go?

**A:** In the **domain service**, NOT in use cases or handlers.

```go
// ✅ Correct - Domain service
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // Business rule: ISBN-13 checksum validation
    // Complex logic here
}

// ✅ Correct - Use case orchestrates
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
    // Orchestration logic
}

// ❌ Wrong - Business logic in use case
func (uc *CreateBookUseCase) Execute(req Request) error {
    if len(req.ISBN) != 13 {  // Business logic doesn't belong here!
        return errors.New("invalid ISBN")
    }
}

// ❌ Wrong - Business logic in handler
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
    if len(req.ISBN) != 13 {  // Definitely not here!
        return
    }
}
```

### Q: What's the difference between domain service and use case?

**A:**

| Domain Service | Use Case |
|---------------|----------|
| **Pure business logic** | **Orchestration** |
| ISBN validation | Validate → Create → Save → Cache |
| Calculate late fees | Get loan → Calculate → Persist |
| Check business rules | Check rule → Update → Notify |
| **No dependencies** | **Uses domain services + repositories** |
| 100% testable | Mocks needed for testing |

### Q: Where should interfaces be defined?

**A:** In the **domain layer**, implemented in adapters.

```go
// ✅ Correct
// internal/domain/book/repository.go
type Repository interface {
    Create(ctx context.Context, book Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
}

// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct { /* ... */ }
func (r *PostgresBookRepository) Create(...) error { /* impl */ }

// ❌ Wrong - Interface in adapter layer
// internal/adapters/repository/postgres/repository.go
type BookRepository interface { /* ... */ }
```

**Why?** Domain defines the contract, adapters fulfill it. Dependency points inward.

### Q: Can domain import from use case layer?

**A:** **NO! Never!** Dependencies must point inward only.

```go
// ❌ FORBIDDEN
// internal/domain/book/service.go
import "library-service/internal/usecase/bookops"  // ← Import cycle!

// ✅ Allowed
// internal/usecase/bookops/create_book.go
import "library-service/internal/domain/book"  // ← OK, points inward
```

**Rule:** Domain → Use Case → Adapters → Infrastructure (one direction only)

## 🧪 Testing Questions

### Q: How do I mock a repository for testing?

**A:** Create a mock implementation of the interface.

```go
// internal/domain/book/mocks/repository.go
type MockRepository struct {
    CreateFunc  func(ctx context.Context, book book.Entity) error
    GetByIDFunc func(ctx context.Context, id string) (book.Entity, error)
}

func (m *MockRepository) Create(ctx context.Context, book book.Entity) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, book)
    }
    return nil
}

// In test:
func TestCreateBook(t *testing.T) {
    mockRepo := &mocks.MockRepository{}
    mockRepo.CreateFunc = func(ctx context.Context, book book.Entity) error {
        return nil  // Or return error to test error handling
    }

    uc := NewCreateBookUseCase(mockRepo, nil)
    // Test...
}
```

### Q: What coverage should I aim for?

**A:**

- **Domain layer:** 100% (critical business logic)
- **Use cases:** 80%+ (orchestration logic)
- **Adapters:** 60%+ (less critical, often integration tested)
- **Overall:** 60%+

```bash
# Check coverage
make test-coverage

# Domain should be 100%
go test -coverprofile=coverage.out ./internal/domain/book/
go tool cover -func=coverage.out
```

### Q: Should I use real database in tests?

**A:** **Not in unit tests.** Use mocks. Use real DB only in integration tests.

```go
// ✅ Unit test - Fast, no database
func TestCreateBookUseCase(t *testing.T) {
    mockRepo := &mocks.MockRepository{/* ... */}
    uc := NewCreateBookUseCase(mockRepo, service)
    // Test business logic
}

// ✅ Integration test - Slower, real database
//go:build integration
func TestBookRepository_Create(t *testing.T) {
    db := testdb.Setup(t)  // Real test database
    repo := postgres.NewBookRepository(db)
    // Test database operations
}
```

## 🔧 Development Questions

### Q: How do I add a new API endpoint?

**A:** Follow these steps in order:

1. **Domain** (if new entity)
2. **Use Case** (business logic orchestration)
3. **Handler** (HTTP layer)
4. **Wire** in container.go
5. **Route** in router.go
6. **Swagger** annotations
7. **Migration** (if database changes)

See `.claude/examples/README.md` for complete code examples.

### Q: How do I debug a failing test?

**A:**

```bash
# Run specific test with verbose output
go test -v -run TestCreateBook ./internal/usecase/bookops/

# Run with race detector
go test -race -run TestCreateBook ./internal/usecase/bookops/

# Run single test multiple times (check for flakiness)
go test -count=10 -run TestCreateBook ./internal/usecase/bookops/

# Clear test cache (if tests are cached)
go clean -testcache

# Debug with delve
dlv test ./internal/usecase/bookops/ -- -test.run TestCreateBook
```

### Q: Port 8080 is already in use. How do I fix it?

**A:**

```bash
# Find and kill process using port 8080
lsof -ti:8080 | xargs kill -9

# Or find the process ID first
lsof -i:8080
# Then kill manually
kill -9 <PID>
```

### Q: How do I reset the database?

**A:**

```bash
# Soft reset (rollback and reapply migrations)
make migrate-down
make migrate-up

# Hard reset (destroy and recreate)
make down
docker volume rm $(docker volume ls -q | grep library)
make up
make migrate-up
```

## 🗃️ Database Questions

### Q: How do I create a migration?

**A:**

```bash
# Create migration files
make migrate-create name=add_ratings_table

# This creates two files:
# migrations/postgres/XXXXXX_add_ratings_table.up.sql
# migrations/postgres/XXXXXX_add_ratings_table.down.sql

# Edit the files, then apply:
make migrate-up
```

### Q: How do I rollback a migration?

**A:**

```bash
# Rollback last migration
make migrate-down

# Rollback multiple migrations
go run cmd/migrate/main.go down
go run cmd/migrate/main.go down
# Repeat as needed
```

### Q: Should I add indexes?

**A:** **Yes**, for:

- Foreign keys (always!)
- Frequently queried columns
- Columns used in WHERE clauses
- Columns used in ORDER BY

```sql
-- Always index foreign keys
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_member_id ON loans(member_id);

-- Index frequently queried columns
CREATE INDEX idx_books_isbn ON books(isbn);

-- Composite indexes for common queries
CREATE INDEX idx_loans_member_status ON loans(member_id, status);
```

## 🔐 Security Questions

### Q: How do I protect an endpoint?

**A:** Add `@Security BearerAuth` to Swagger annotations and use auth middleware.

```go
// In handler:
// @Security BearerAuth
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // Handler code
}

// In router.go:
r.Route("/books", func(r chi.Router) {
    r.Use(authMiddleware)  // ← Enforces authentication
    r.Post("/", handlers.Book.CreateBook)
})
```

### Q: How do I get a JWT token for testing?

**A:**

```bash
# Register (or login if already registered)
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# Use the token
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/books
```

### Q: How do I prevent SQL injection?

**A:** **Always** use parameterized queries with `$1, $2` placeholders.

```go
// ✅ Correct - Parameterized query
query := "SELECT * FROM books WHERE id = $1"
row := r.db.QueryRowContext(ctx, query, bookID)

// ✅ Correct - Multiple parameters
query := "INSERT INTO books (id, name, isbn) VALUES ($1, $2, $3)"
_, err := r.db.ExecContext(ctx, query, book.ID, book.Name, book.ISBN)

// ❌ DANGEROUS - String concatenation
query := "SELECT * FROM books WHERE id = '" + bookID + "'"  // SQL INJECTION!

// ❌ DANGEROUS - fmt.Sprintf
query := fmt.Sprintf("SELECT * FROM books WHERE id = '%s'", bookID)  // NO!
```

## 📝 Swagger / API Questions

### Q: Swagger UI shows "Unauthorized" for all endpoints

**A:** Click the "Authorize" button and enter: `Bearer <your-token>`

```
1. Get token: See "How do I get a JWT token" above
2. Click "Authorize" button in Swagger UI (top right)
3. Enter: Bearer eyJhbGc...  (include "Bearer " prefix!)
4. Click "Authorize"
5. Try endpoint again
```

### Q: My Swagger changes aren't showing

**A:** Regenerate Swagger docs:

```bash
make gen-docs

# Then restart API
make run
```

### Q: What Swagger annotations are required?

**A:** At minimum:

```go
// @Summary      Brief description (REQUIRED)
// @Tags         category
// @Security     BearerAuth (for protected endpoints!)
// @Param        name type dataType required "description"
// @Success      200 {object} dto.Response
// @Failure      400 {object} dto.ErrorResponse
// @Router       /path [method] (REQUIRED)
```

## 🚀 Performance Questions

### Q: How do I profile my code?

**A:**

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/domain/book/
go tool pprof cpu.prof
# Then type: top, list <function>, web

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/domain/book/
go tool pprof mem.prof
```

### Q: How do I fix N+1 query problems?

**A:** Use eager loading or batch queries.

```go
// ❌ N+1 Problem
books, _ := repo.List(ctx)
for _, book := range books {
    authors, _ := authorRepo.GetByBookID(ctx, book.ID)  // N queries!
}

// ✅ Solution: Eager load
books, _ := repo.ListWithAuthors(ctx)  // Single query with JOIN
```

### Q: Should I add caching?

**A:** Add caching for:

- Frequently accessed data (hot data)
- Rarely changing data
- Expensive computations

```go
// Use case with cache
func (uc *GetBookUseCase) Execute(ctx context.Context, id string) (*book.Entity, error) {
    // Check cache first
    if cached, err := uc.cache.Get(ctx, id); err == nil {
        return cached, nil
    }

    // Not in cache, fetch from DB
    book, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache
    uc.cache.Set(ctx, id, book)

    return &book, nil
}
```

## 🐛 Common Errors

### Q: "import cycle not allowed"

**A:** You're violating dependency direction. Domain is importing from outer layers.

**Fix:** Move interface to domain, implement in adapter.

### Q: "cannot find package"

**A:**

```bash
go mod tidy
go mod download
go clean -modcache
go build ./...
```

### Q: Tests fail in CI but pass locally

**A:** Common causes:

1. **Race condition** → Run `go test -race ./...`
2. **Time-dependent** → Don't use `time.Now()` in tests
3. **Environment** → Check environment variables
4. **Test cache** → CI clears cache, you might have stale cache

### Q: "pq: relation does not exist"

**A:** Migrations not run.

```bash
make migrate-up

# Or check migration status
docker exec -it $(docker ps -qf "name=postgres") \
    psql -U library -d library -c "SELECT * FROM schema_migrations;"
```

## 📚 Documentation Questions

### Q: Where do I find code examples?

**A:** `.claude/examples/README.md` - Complete working examples for all layers.

### Q: Where do I find quick commands?

**A:** `.claude/recipes.md` - Copy-paste solutions for common tasks.

### Q: Where do I find troubleshooting help?

**A:** `.claude/troubleshooting.md` - Solutions to common problems.

### Q: What should I read first as a new developer?

**A:** Follow this order:

1. `.claude/onboarding.md` (15 minutes)
2. `.claude/cheatsheet.md` (quick reference)
3. `.claude/examples/` (see actual code)
4. `.claude/architecture.md` (deep dive)

## 🔄 Workflow Questions

### Q: What's the complete workflow for adding a feature?

**A:**

```bash
# 1. Create feature branch
git checkout -b feature/add-loans

# 2. Implement (follow layer order)
# - Domain layer (entity, service, repository interface, tests)
# - Use case layer (create use cases with "ops" suffix)
# - Adapter layer (repository impl, handlers, DTOs)
# - Wire in container.go
# - Add routes in router.go
# - Create migration if needed

# 3. Test
make test

# 4. Update docs
make gen-docs

# 5. Run full CI
make ci

# 6. Commit
git add .
git commit -m "feat: add loan management"

# 7. Push
git push origin feature/add-loans
```

### Q: What should I check before committing?

**A:** Use `.claude/checklist.md` or run:

```bash
make ci  # Runs: fmt → vet → lint → test → build
```

If `make ci` passes, you're good to commit!

## 💡 Best Practices Questions

### Q: How should I name my tests?

**A:** `Test<Function>_<Scenario>_<ExpectedResult>`

```go
func TestService_ValidateISBN_ValidISBN13_ReturnsNoError(t *testing.T)
func TestService_ValidateISBN_InvalidChecksum_ReturnsError(t *testing.T)
func TestCreateBookUseCase_Execute_DuplicateISBN_ReturnsAlreadyExistsError(t *testing.T)
```

### Q: How should I handle errors?

**A:** Always wrap with context using `%w`:

```go
// ✅ Good
if err := repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}

// ✅ Good - can check with errors.Is()
if errors.Is(err, errors.ErrNotFound) {
    // Handle
}

// ❌ Bad - loses context
if err := repo.Create(ctx, book); err != nil {
    return err
}
```

### Q: Should I use panic()?

**A:** **No!** Return errors instead.

```go
// ❌ Bad
if book == nil {
    panic("book is nil")
}

// ✅ Good
if book == nil {
    return errors.New("book cannot be nil")
}
```

## 🎯 Quick Answers

| Question | Answer |
|----------|--------|
| Where does business logic go? | Domain service |
| Where does orchestration go? | Use case |
| Where are interfaces defined? | Domain layer |
| Where are interfaces implemented? | Adapter layer |
| What's the dependency direction? | Domain ← Use Case ← Adapters ← Infrastructure |
| Use case package naming? | Add "ops" suffix (bookops, authops) |
| Required test coverage? | Domain 100%, Use Case 80%, Overall 60% |
| How to debug a test? | `go test -v -run TestName ./path/` |
| How to reset database? | `make down && make up && make migrate-up` |
| How to get JWT token? | POST to `/auth/login` with email/password |
| Port in use? | `lsof -ti:8080 | xargs kill -9` |
| Swagger not updating? | `make gen-docs` then restart API |

---

**Still have questions? Check `.claude/` documentation or ask with specific context!** 💬
