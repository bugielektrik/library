# Common Tasks - Quick Reference

Frequently asked questions and quick solutions for common development tasks.

**Format:** Quick answer first, then command/code if needed.

---

## Database & Migrations

### How do I create a new migration?

```bash
make migrate-create name=your_migration_name
# Creates migrations/postgres/{timestamp}_your_migration_name.up.sql and .down.sql
```

Edit both files, then apply:
```bash
make migrate-up
```

### How do I rollback the last migration?

```bash
make migrate-down  # Rolls back 1 migration
```

### How do I reset the database completely?

```bash
# WARNING: Destroys all data
make migrate-down  # Rollback all migrations
make migrate-up    # Reapply all migrations
```

### What's the database connection string format?

```
postgres://username:password@host:port/database?sslmode=disable
```

**Development default:**
```
postgres://library:library123@localhost:5432/library?sslmode=disable
```

Set via environment variable:
```bash
export POSTGRES_DSN="postgres://..."
```

---

## Testing

### How do I run only unit tests (fast)?

```bash
make test-unit
# Or:
go test ./... -short
```

### How do I run tests for a specific package?

```bash
# Specific package
go test ./internal/domain/book/...

# Specific use case
go test ./internal/usecase/bookops/...

# Specific test function
go test ./internal/domain/book/... -run TestValidateISBN
```

### How do I run integration tests?

```bash
# Requires PostgreSQL running
make up
make migrate-up
make test-integration

# Or manually:
go test ./... -tags=integration -v
```

### How do I get test coverage?

```bash
make test-coverage
# Opens HTML report in browser
```

### How do I clear test cache?

```bash
go clean -testcache
```

---

## Code Quality

### How do I format my code?

```bash
make fmt
# Runs: gofmt + goimports
```

### How do I run the linter?

```bash
make lint
# Runs: golangci-lint with project config
```

### How do I run the full CI pipeline locally?

```bash
make ci
# Runs: fmt → vet → lint → test → build
```

### How do I check for unused code?

```bash
# Included in make lint, or run directly:
golangci-lint run --enable unused
```

---

## Building & Running

### How do I build the project?

```bash
# Build all binaries
make build

# Build specific binary
make build-api      # → bin/library-api
make build-worker   # → bin/library-worker
make build-migrate  # → bin/library-migrate
```

### How do I run the API server locally?

```bash
# Full development stack (recommended)
make dev

# Or manually:
make up              # Start PostgreSQL + Redis
make migrate-up      # Apply migrations
make run             # Run API server
```

### How do I stop the development stack?

```bash
make down
# Stops PostgreSQL + Redis containers
```

### How do I check if the server is running?

```bash
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

---

## API Development

### How do I add a new API endpoint?

**Quick version:**
1. Create use case in `internal/usecase/{entity}ops/`
2. Add handler method in `internal/adapters/http/handlers/`
3. Add route in handler's `Routes()` method
4. Wire use case in `internal/usecase/container.go`
5. Add Swagger annotations
6. Run `make gen-docs`

**See:** [Adding API Endpoint Guide](./adding-api-endpoint.md)

### How do I regenerate Swagger documentation?

```bash
make gen-docs
# View at: http://localhost:8080/swagger/index.html
```

### How do I test an endpoint with curl?

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}'

# Login (get token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}'

# Use token
TOKEN="your_access_token_here"
curl -X GET http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN"
```

### How do I add validation to a DTO?

Add struct tags to DTO fields:

```go
type CreateBookRequest struct {
    Name    string   `json:"name" validate:"required,min=1,max=255"`
    Genre   string   `json:"genre" validate:"required,min=1,max=100"`
    ISBN    string   `json:"isbn" validate:"required,isbn"`
    Authors []string `json:"authors" validate:"required,min=1,dive,uuid4"`
}
```

**Common validation tags:**
- `required` - Field must be present
- `min=N,max=N` - String length or number range
- `email` - Valid email format
- `uuid4` - Valid UUID v4
- `isbn` - Valid ISBN format
- `dive` - Validate slice elements
- `omitempty` - Skip validation if empty

**See:** https://pkg.go.dev/github.com/go-playground/validator/v10

---

## Error Handling

### How do I return a domain error from a use case?

Use predefined errors from `pkg/errors`:

```go
import "library-service/pkg/errors"

// Not found
return nil, errors.ErrNotFound.WithDetails("entity", "book")

// Already exists
return nil, errors.ErrAlreadyExists.WithDetails("field", "email")

// Validation error
return nil, errors.ErrValidation.WithMessage("ISBN format is invalid")

// Generic error
return nil, errors.ErrInternal.WithMessage("failed to process payment")
```

### How do I wrap errors with context?

Use `fmt.Errorf` with `%w`:

```go
if err := s.repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}
```

**Don't do this:**
```go
// Bad: loses error type information
return errors.New("creating book: " + err.Error())
```

### What HTTP status codes map to domain errors?

- `ErrNotFound` → 404 Not Found
- `ErrAlreadyExists` → 409 Conflict
- `ErrValidation` → 400 Bad Request
- `ErrUnauthorized` → 401 Unauthorized
- `ErrForbidden` → 403 Forbidden
- `ErrInternal` → 500 Internal Server Error

---

## Logging

### How do I add logging to a use case?

```go
import (
    "go.uber.org/zap"
    "library-service/pkg/logutil"
)

func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*CreateBookResponse, error) {
    logger := logutil.UseCaseLogger(ctx, "create_book",
        zap.String("name", req.Name),
        zap.String("isbn", req.ISBN),
    )

    // Validation failure (business error)
    if err := uc.service.ValidateISBN(req.ISBN); err != nil {
        logger.Warn("validation failed", zap.Error(err))
        return nil, err
    }

    // System error
    if err := uc.repo.Create(ctx, book); err != nil {
        logger.Error("failed to create book", zap.Error(err))
        return nil, err
    }

    // Success
    logger.Info("book created", zap.String("id", book.ID))
    return &CreateBookResponse{Book: book}, nil
}
```

### How do I add logging to a handler?

```go
import (
    "go.uber.org/zap"
    "library-service/pkg/logutil"
)

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logutil.HandlerLogger(ctx, "book_handler", "create")

    // ... handler logic ...

    logger.Info("book created via API", zap.String("id", response.ID))
    h.RespondJSON(w, http.StatusCreated, response)
}
```

### What log level should I use?

- **Debug**: Detailed internal state (rarely used)
- **Info**: Successful operations, key events
- **Warn**: Business validation failures, expected errors
- **Error**: System failures, unexpected errors
- **Fatal**: Application cannot continue (avoid in libraries)

---

## Dependency Injection

### How do I add a new repository to the container?

**Step 1:** Add to `Repositories` struct in `internal/usecase/container.go`:

```go
type Repositories struct {
    BookRepo        book.Repository
    MemberRepo      member.Repository
    YourNewRepo     yournew.Repository  // ADD THIS
}
```

**Step 2:** Wire in application bootstrap `internal/infrastructure/app/app.go`:

```go
repos := usecase.Repositories{
    BookRepo:    postgres.NewBookRepository(db),
    MemberRepo:  postgres.NewMemberRepository(db),
    YourNewRepo: postgres.NewYourNewRepository(db),  // ADD THIS
}
```

### How do I add a new use case to the container?

**Step 1:** Add to `Container` struct in `internal/usecase/container.go`:

```go
type Container struct {
    CreateBook      *bookops.CreateBookUseCase
    YourNewUseCase  *yourops.YourNewUseCase  // ADD THIS
}
```

**Step 2:** Wire in `NewContainer()` function:

```go
func NewContainer(repos Repositories, caches Caches, authServices AuthServices) *Container {
    // Create domain services
    bookService := book.NewService()
    yourService := yournew.NewService()  // ADD THIS

    return &Container{
        CreateBook: bookops.NewCreateBookUseCase(
            repos.BookRepo,
            caches.BookCache,
            bookService,
        ),
        YourNewUseCase: yourops.NewYourNewUseCase(
            repos.YourNewRepo,
            yourService,  // ADD THIS
        ),
    }
}
```

### What's the difference between a domain service and a use case?

**Domain Service** (`internal/domain/{entity}/service.go`):
- Pure business logic (validation, calculations)
- NO external dependencies (no DB, HTTP, frameworks)
- Created in `NewContainer()` function
- Example: `book.NewService()`, `member.NewService()`

**Use Case** (`internal/usecase/{entity}ops/{operation}.go`):
- Orchestrates domain entities and services
- Calls repositories, caches, external services
- Created in `NewContainer()` and added to `Container` struct
- Example: `CreateBookUseCase`, `LoginMemberUseCase`

---

## Architecture Questions

### Where should I put business logic?

**Domain layer** (`internal/domain/{entity}/`):
- Entity validation rules
- Business calculations
- Domain-specific errors
- Pure functions (no I/O)

**Example:**
```go
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // ISBN validation logic
}
```

### Where should I put orchestration logic?

**Use case layer** (`internal/usecase/{entity}ops/`):
- Coordinating multiple domain entities
- Transaction management
- Calling repositories
- Calling external services

**Example:**
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (*CreateBookResponse, error) {
    // 1. Validate with domain service
    // 2. Create entity
    // 3. Save to repository
    // 4. Update cache
    // 5. Return response
}
```

### Where should I put HTTP-specific logic?

**Handler layer** (`internal/adapters/http/handlers/`):
- Request decoding
- Response encoding
- HTTP status codes
- Request validation (DTO validation)

**Example:**
```go
// internal/adapters/http/handlers/book_crud.go
func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    // 1. Decode JSON
    // 2. Validate DTO
    // 3. Call use case
    // 4. Return HTTP response
}
```

### Why are use case packages named with "ops" suffix?

To avoid naming conflicts with domain packages:

```go
import (
    "library-service/internal/domain/book"      // package book
    "library-service/internal/usecase/bookops"  // package bookops
)

// Clean references without aliases
bookEntity := book.New(...)
useCase := bookops.NewCreateBookUseCase(...)
```

**Without "ops" suffix:**
```go
// Bad: requires aliases
import (
    domainbook "library-service/internal/domain/book"
    usecasebook "library-service/internal/usecase/book"
)
```

---

## Docker & Deployment

### How do I start the Docker development stack?

```bash
make up
# Starts: PostgreSQL (port 5432) + Redis (port 6379)
```

### How do I view Docker container logs?

```bash
cd deployments/docker
docker-compose logs -f postgres
docker-compose logs -f redis
```

### How do I connect to the PostgreSQL container?

```bash
psql -h localhost -U library -d library
# Password: library123
```

### How do I rebuild Docker images?

```bash
cd deployments/docker
docker-compose down
docker-compose up --build
```

---

## Git & Version Control

### How do I create a new feature branch?

```bash
git checkout -b feature/your-feature-name
```

### What's the commit message format?

Follow Conventional Commits:

```
type(scope): brief description

Longer description if needed.

- Bullet points for details
- Multiple changes

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code restructuring
- `docs:` - Documentation
- `test:` - Tests
- `chore:` - Maintenance

**Examples:**
```
feat(book): add availability check endpoint
fix(auth): handle expired refresh tokens correctly
refactor(handlers): split book handler into separate files
docs(examples): add integration testing guide
```

### How do I run tests before committing?

```bash
make ci
# Runs: fmt → vet → lint → test → build
```

Add to `.git/hooks/pre-commit` for automatic checks:
```bash
#!/bin/bash
make ci
```

---

## Troubleshooting

### Port 8080 is already in use

```bash
# Find process using port 8080
lsof -ti:8080

# Kill process
lsof -ti:8080 | xargs kill -9

# Or use different port
export PORT=8081
make run
```

### Database connection refused

```bash
# Check if PostgreSQL is running
make up
docker ps | grep postgres

# Check connection string
echo $POSTGRES_DSN

# Test connection
psql -h localhost -U library -d library
```

### Tests fail with "database does not exist"

```bash
# Create test database
psql -h localhost -U library -d library -c "CREATE DATABASE library_test;"

# Run migrations
POSTGRES_DSN="postgres://library:library123@localhost:5432/library_test?sslmode=disable" \
  go run cmd/migrate/main.go up
```

### "command not found" errors

```bash
# Install development tools
make install-tools

# Verify installation
which golangci-lint
which swag
```

### Import cycle errors

Check dependency order: **Domain → Use Case → Adapters → Infrastructure**

**Common mistake:** Domain importing from use case or adapters.

**Fix:** Move shared types to domain layer.

---

## Performance

### How do I run benchmarks?

```bash
make benchmark

# Or specific package
go test -bench=. ./internal/domain/book/...

# With memory allocation stats
go test -bench=. -benchmem ./...
```

### How do I profile the application?

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### How do I check for memory leaks?

```bash
# Run with race detector
go test -race ./...

# Run with memory sanitizer
go test -msan ./...
```

---

## Quick Reference

### Project Structure
```
internal/
├── domain/              # Business logic (pure, no dependencies)
├── usecase/            # Application orchestration (depends on domain)
├── adapters/           # External interfaces (HTTP, DB, cache)
└── infrastructure/     # Technical concerns (config, logging, server)
```

### Key Files
- `Makefile` - All commands (30+ targets)
- `.golangci.yml` - Linter configuration
- `internal/usecase/container.go` - Dependency injection
- `internal/infrastructure/app/app.go` - Application bootstrap
- `cmd/api/main.go` - API entry point + Swagger metadata

### Essential Commands
```bash
make init              # First-time setup
make dev               # Full development stack
make test              # Run all tests
make ci                # Full CI pipeline
make gen-docs          # Regenerate Swagger docs
```

### Need More Help?

- **Adding Features:** See [adding-domain-entity.md](./adding-domain-entity.md), [adding-api-endpoint.md](./adding-api-endpoint.md)
- **Testing:** See [integration-testing.md](./integration-testing.md)
- **Architecture:** See [.claude/architecture.md](../.claude/architecture.md)
- **Full Documentation:** See [.claude/README.md](../.claude/README.md)
