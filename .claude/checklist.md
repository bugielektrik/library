# Code Review Checklist

> **Pre-commit checklist for maintaining code quality**

## ⚡ Quick Pre-Commit

Run this before every commit:

```bash
make ci  # Runs: fmt → vet → lint → test → build
```

If `make ci` passes, you're 90% ready. Use this checklist for the final 10%.

## 📋 Essential Checks

### ✅ Code Quality

- [ ] **All tests pass** (`make test`)
- [ ] **No linter errors** (`make lint`)
- [ ] **Code is formatted** (`make fmt`)
- [ ] **No race conditions** (`go test -race ./...`)
- [ ] **Coverage maintained** (check `make test-coverage`)

### ✅ Architecture

- [ ] **Dependencies point inward**: Domain ← Use Case ← Adapters ← Infrastructure
- [ ] **No domain → outer layer imports** (domain must be pure)
- [ ] **Package naming correct**: Use cases have "ops" suffix (e.g., `bookops`)
- [ ] **Business logic in domain**, not handlers
- [ ] **Handlers are thin**, delegate to use cases
- [ ] **Interfaces defined in domain**, implemented in adapters

### ✅ Error Handling

- [ ] **Errors wrapped with context**: `fmt.Errorf("operation: %w", err)`
- [ ] **Domain errors used**: `errors.ErrNotFound`, `errors.ErrValidation`
- [ ] **All errors checked** (no ignored errors without `// nolint:errcheck`)
- [ ] **Errors logged appropriately**: `log.Error("msg", "error", err)`

### ✅ Testing

- [ ] **Unit tests for domain layer** (aim for 100%)
- [ ] **Use case tests with mocks**
- [ ] **Table-driven tests** where applicable
- [ ] **Tests follow naming**: `TestFunction_Scenario_ExpectedResult`
- [ ] **No tests skipped** (no `t.Skip()` without good reason)
- [ ] **Test fixtures cleaned up** (no stale test data)

### ✅ API / HTTP

- [ ] **Swagger annotations complete**
  - [ ] `@Summary` present
  - [ ] `@Security BearerAuth` for protected endpoints
  - [ ] `@Param` for all parameters
  - [ ] `@Success` and `@Failure` responses
  - [ ] `@Router` path and method
- [ ] **DTOs have validation tags** (`validate:"required,email"`, etc.)
- [ ] **Status codes appropriate** (201 for create, 204 for delete, etc.)
- [ ] **Swagger docs regenerated** (`make gen-docs`)

### ✅ Database

- [ ] **Migrations have both up and down**
- [ ] **Migration names descriptive**: `create_loans_table` not `migration1`
- [ ] **Indexes added** for foreign keys and frequently queried columns
- [ ] **Constraints defined**: `NOT NULL`, `UNIQUE`, `CHECK`, etc.
- [ ] **SQL injection prevented** (use parameterized queries)
- [ ] **Transactions used** where needed

### ✅ Security

- [ ] **No hardcoded secrets** (check for passwords, API keys, tokens)
- [ ] **No `.env` file committed**
- [ ] **Authentication enforced** on protected routes
- [ ] **Input validated** before processing
- [ ] **SQL injection prevented** (always use `$1, $2` placeholders)
- [ ] **XSS prevented** (proper escaping in responses)

### ✅ Documentation

- [ ] **Public functions have comments**: Start with function name
- [ ] **Complex logic explained**: Why, not what
- [ ] **TODOs have owner**: `// TODO(username): ...`
- [ ] **README updated** if needed
- [ ] **API docs current** (Swagger regenerated)

### ✅ Git

- [ ] **Commit message clear**: `feat:`, `fix:`, `refactor:`, etc.
- [ ] **No large files committed** (check `git diff --stat`)
- [ ] **No debug code** (`console.log`, `fmt.Println`, etc.)
- [ ] **Branch up to date** with main
- [ ] **.gitignore updated** if new files added

## 🎯 Layer-Specific Checks

### Domain Layer

- [ ] **Zero external dependencies** (no imports from adapters/infrastructure)
- [ ] **Pure business logic only**
- [ ] **Interfaces defined here**, implemented elsewhere
- [ ] **100% test coverage** (critical!)
- [ ] **No database, HTTP, or framework code**
- [ ] **Entity validation** in domain service
- [ ] **Immutability** where possible (entity fields)

**Red Flags:**
```go
// ❌ Domain importing from outer layers
import "library-service/internal/adapters/repository/postgres"
import "library-service/internal/usecase/bookops"

// ❌ HTTP in domain
func (s *Service) HandleRequest(w http.ResponseWriter, r *http.Request)

// ❌ Database in domain
func (s *Service) Query(db *sql.DB)
```

### Use Case Layer

- [ ] **Depends only on domain interfaces**
- [ ] **One use case = one file**
- [ ] **Returns domain entities**, not DTOs
- [ ] **Orchestration only**, no business logic
- [ ] **Context as first parameter**: `Execute(ctx context.Context, ...)`
- [ ] **Error wrapping**: `fmt.Errorf("context: %w", err)`

**Red Flags:**
```go
// ❌ Business logic in use case
func (uc *CreateBookUseCase) Execute(...) {
    if len(book.ISBN) != 13 {  // Should be in domain service!
        return errors.New("invalid ISBN")
    }
}

// ❌ Returning DTOs
func (uc *GetBookUseCase) Execute(...) (*dto.BookResponse, error) {
    // Should return domain.Entity
}
```

### Adapter Layer

- [ ] **Thin layer**, implements domain interfaces
- [ ] **DTOs for external format** (JSON, XML, etc.)
- [ ] **No business logic**
- [ ] **Error mapping** (database errors → domain errors)
- [ ] **SQL parameterized** (never string concatenation)

**Red Flags:**
```go
// ❌ Business logic in adapter
func (r *Repository) Create(book domain.Entity) error {
    if book.Price > 100 {  // Business logic! Belongs in domain
        return errors.New("too expensive")
    }
}

// ❌ SQL injection risk
query := "SELECT * FROM books WHERE id = '" + id + "'"  // DANGEROUS!

// ✅ Correct
query := "SELECT * FROM books WHERE id = $1"
```

### Handler Layer

- [ ] **Thin handlers**, delegate to use cases
- [ ] **Validate input** before calling use case
- [ ] **Map between DTOs and domain entities**
- [ ] **Appropriate status codes**
- [ ] **Error responses formatted** consistently
- [ ] **Swagger annotations complete**

**Red Flags:**
```go
// ❌ Business logic in handler
func (h *Handler) CreateBook(...) {
    if len(req.ISBN) == 10 || len(req.ISBN) == 13 {  // Belongs in domain!
        // ...
    }
}

// ❌ Direct database access
func (h *Handler) GetBook(...) {
    row := h.db.QueryRow("SELECT * FROM books WHERE id = $1", id)  // Wrong layer!
}
```

## 📊 Performance Checks

- [ ] **N+1 queries avoided** (use eager loading)
- [ ] **Indexes exist** for frequently queried columns
- [ ] **Pagination implemented** for list endpoints
- [ ] **Connection pooling configured**
- [ ] **Slow queries identified** (check with `EXPLAIN ANALYZE`)
- [ ] **Benchmarks run** if performance-critical (`go test -bench=.`)

## 🧪 Test Quality

### Unit Test Checklist

- [ ] **Tests are fast** (< 1 second for unit tests)
- [ ] **Tests are isolated** (no shared state)
- [ ] **No database in unit tests** (use mocks)
- [ ] **No external APIs** (mock them)
- [ ] **Deterministic** (no random values, time.Now())
- [ ] **Descriptive names**: `TestService_ValidateISBN_InvalidChecksum`
- [ ] **Table-driven** where multiple scenarios exist

```go
// ✅ Good unit test
func TestValidateISBN(t *testing.T) {
    svc := NewService()  // No dependencies

    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "9780132350884", false},
        {"invalid checksum", "9780132350881", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Test Checklist

- [ ] **Build tag**: `//go:build integration`
- [ ] **Real database** (test database, not production!)
- [ ] **Cleanup after test** (`defer teardown()`)
- [ ] **Isolated** (can run in parallel)
- [ ] **Idempotent** (can run multiple times)

## 🚨 Common Issues

### Import Cycles

**Check:**
```bash
go build ./...
# If "import cycle not allowed" → fix dependency direction
```

**Fix:**
- Move interface to domain
- Remove outer layer imports from domain
- Use dependency injection

### Memory Leaks

**Check:**
```bash
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

**Common causes:**
- Unclosed database connections
- Goroutine leaks (no cleanup)
- Large in-memory caches

### Race Conditions

**Check:**
```bash
go test -race ./...
```

**Common causes:**
- Shared maps without mutex
- Concurrent writes to slices
- Unprotected counter variables

## 📝 Before Pushing

```bash
# 1. Update dependencies
go mod tidy

# 2. Run full CI
make ci

# 3. Check for issues
golangci-lint run

# 4. Verify Swagger
make gen-docs

# 5. Run integration tests
make test-integration

# 6. Check git status
git status

# 7. Review diff
git diff --stat
git diff

# 8. Commit
git add .
git commit -m "feat: descriptive message"

# 9. Push
git push origin feature-branch
```

## 🎯 Quick Self-Review Script

Save as `.claude/scripts/review.sh`:

```bash
#!/bin/bash
set -e

echo "🔍 Running self-review checks..."

echo "✓ Formatting code..."
make fmt

echo "✓ Running vet..."
make vet

echo "✓ Running linter..."
make lint

echo "✓ Running tests..."
make test

echo "✓ Checking race conditions..."
go test -race ./...

echo "✓ Checking for TODO without owner..."
if grep -r "TODO:" --include="*.go" . | grep -v "TODO([a-z]*):" ; then
    echo "❌ Found TODOs without owner. Use: // TODO(username): description"
    exit 1
fi

echo "✓ Checking for debug statements..."
if grep -r "fmt.Println\|log.Println" --include="*.go" ./internal ; then
    echo "⚠️  Warning: Found debug print statements"
fi

echo "✓ Checking for hardcoded secrets..."
if grep -ri "password.*=.*\"" --include="*.go" . | grep -v "_test.go" ; then
    echo "❌ Possible hardcoded password found!"
    exit 1
fi

echo "✓ Checking migrations..."
if ls migrations/postgres/*.up.sql 1> /dev/null 2>&1; then
    for up in migrations/postgres/*.up.sql; do
        down="${up%.up.sql}.down.sql"
        if [ ! -f "$down" ]; then
            echo "❌ Missing down migration for $up"
            exit 1
        fi
    done
fi

echo "✓ Building..."
make build

echo "✅ All checks passed! Ready to commit."
```

Make it executable:
```bash
chmod +x .claude/scripts/review.sh
```

Run before commit:
```bash
.claude/scripts/review.sh
```

## 📚 References

- Architecture: `.claude/architecture.md`
- Standards: `.claude/standards.md`
- Testing: `.claude/testing.md`
- Examples: `.claude/examples/`

---

**Remember: Code is read more often than written. Make it clear, not clever.** 💡
