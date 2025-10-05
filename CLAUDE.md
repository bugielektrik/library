# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **📚 Full Documentation:** See [`.claude`](./.claude/) directory for comprehensive guides

## Project Overview

Library Management System - A Go-based REST API following Clean Architecture principles, optimized for vibecoding with Claude Code. The system manages books, authors, members, and subscriptions with JWT authentication.

**Key Technologies:** Go 1.25, PostgreSQL, Redis, Chi router, JWT, Docker, Swagger/OpenAPI

## Documentation Index

- **[Quick Start & Navigation](./.claude/README.md)** - 30-second quick reference
- **[Commands Reference](./.claude/commands.md)** - All essential commands
- **[Setup Guide](./.claude/setup.md)** - First-time setup and troubleshooting
- **[Architecture](./.claude/architecture.md)** - Clean architecture patterns
- **[Development Workflow](./.claude/development.md)** - Daily development tasks
- **[Testing Guide](./.claude/testing.md)** - Testing patterns and strategies
- **[API Documentation](./.claude/api.md)** - REST API endpoints
- **[Code Standards](./.claude/standards.md)** - Go best practices and conventions

## Architecture (Clean Architecture)

The codebase follows strict dependency rules: **Domain → Use Case → Adapters → Infrastructure**

```
internal/
├── domain/              # Business logic (ZERO external dependencies)
│   ├── book/           # Book entity, service, repository interface
│   ├── member/         # Member entity, service (subscriptions)
│   └── author/         # Author entity
├── usecase/            # Application orchestration (depends on domain)
│   ├── bookops/        # CreateBook, UpdateBook, etc. ("ops" suffix)
│   ├── authops/        # Register, Login, RefreshToken ("ops" suffix)
│   └── subops/         # SubscribeMember ("ops" suffix)
├── adapters/           # External interfaces (HTTP, DB, cache)
│   ├── http/           # Chi handlers, middleware, DTOs
│   ├── repository/     # PostgreSQL/MongoDB/Memory implementations
│   └── cache/          # Redis/Memory cache implementations
└── infrastructure/     # Technical concerns
    ├── auth/           # JWT token generation/validation
    ├── store/          # Database connections
    └── server/         # HTTP server configuration
```

**Critical Rules:**
- Domain layer must NEVER import from outer layers
- Use case packages use "ops" suffix (e.g., `bookops`) to avoid naming conflicts with domain packages (e.g., `book`)
- Use cases define behavior via interfaces, adapters provide implementations

## Common Commands

### Building
```bash
make build              # Build all binaries (api, worker, migrate)
make build-api          # Build API server only → bin/library-api
make build-worker       # Build worker only → bin/library-worker
make build-migrate      # Build migration tool → bin/library-migrate
```

### Running Locally
```bash
# Full stack (recommended for development)
make dev                # Starts docker services + migrations + API server

# Individual services
make run                # Run API server (requires PostgreSQL/Redis running)
make run-worker         # Run background worker
make up                 # Start docker-compose (PostgreSQL + Redis)
make down               # Stop docker services

# Quick start (5 minutes)
make init && make up && make migrate-up && make run
```

### Testing
```bash
make test               # All tests with race detection + coverage
make test-unit          # Unit tests only (fast, no database)
make test-integration   # Integration tests (requires database)
make test-coverage      # Generate HTML coverage report

# Run specific package tests
go test -v ./internal/domain/book/...
go test -v -run TestCreateBook ./internal/usecase/bookops/
```

### Code Quality
```bash
make ci                 # Full CI pipeline: fmt → vet → lint → test → build
make lint               # Run golangci-lint (25+ linters enabled)
make fmt                # Format code with gofmt + goimports
make vet                # Run go vet for suspicious constructs
```

### Database Migrations
```bash
make migrate-up         # Apply all pending migrations
make migrate-down       # Rollback last migration
make migrate-create name=add_book_ratings  # Create new migration

# Direct usage (requires POSTGRES_DSN env var)
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
POSTGRES_DSN="postgres://library:library123@localhost:5432/library?sslmode=disable" go run cmd/migrate/main.go up
```

### Development Tools
```bash
make install-tools      # Install golangci-lint, mockgen, swag
make gen-mocks          # Generate test mocks
make gen-docs           # Generate Swagger/OpenAPI docs (see API Documentation section)
make benchmark          # Run performance benchmarks
```

## API Documentation

**Swagger UI:** http://localhost:8080/swagger/index.html (when server is running)

**Regenerating Swagger Documentation:**
```bash
# Full regeneration (recommended)
make gen-docs

# Manual regeneration with dependency parsing
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal
```

**Adding API Documentation to Handlers:**
```go
// @Summary Create a new book
// @Description Create a new book with the provided details
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateBookRequest true "Book details"
// @Success 201 {object} dto.BookResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // Handler implementation
}
```

**Important Swagger Annotations:**
- `@Summary` - Brief description (required)
- `@Description` - Detailed explanation
- `@Tags` - Group endpoints together
- `@Security BearerAuth` - **Required for all protected endpoints** (all /books routes, /auth/me)
- `@Param` - Request parameters (body, path, query, header)
- `@Success` / `@Failure` - Response codes with schemas
- `@Router` - Endpoint path and HTTP method

**Security Definition:**
The JWT security scheme is defined in `cmd/api/main.go`:
```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
```

**Testing Authentication in Swagger UI:**
1. Register/Login to get JWT token
2. Click "Authorize" button in Swagger UI
3. Enter: `Bearer <your-access-token>`
4. All protected endpoints will now include the Authorization header

## Development Workflow

### Adding a New Feature

**Example: Adding a "Loan" domain**

1. **Domain Layer** (business logic first):
   ```bash
   # Create domain structure
   mkdir -p internal/domain/loan
   touch internal/domain/loan/{entity.go,service.go,repository.go,dto.go}
   ```
   - Define `Loan` entity with business rules
   - Create `LoanService` for complex logic (e.g., overdue fees, borrowing limits)
   - Define `Repository` interface (NOT implementation)
   - Write unit tests with 100% coverage

2. **Use Case Layer** (orchestration):
   ```bash
   mkdir -p internal/usecase/loanops  # Note: "ops" suffix
   touch internal/usecase/loanops/{create_loan.go,return_loan.go}
   ```
   - Create use cases that orchestrate domain services
   - Each use case = one file (single responsibility)
   - Mock repositories in tests

3. **Adapter Layer** (HTTP + DB):
   ```bash
   # Repository implementation
   touch internal/adapters/repository/postgres/loan.go

   # HTTP handlers
   touch internal/adapters/http/handlers/loan.go
   touch internal/adapters/http/dto/loan.go
   ```
   - Implement repository interface for PostgreSQL
   - Create HTTP handlers (thin layer, delegate to use cases)
   - Add DTOs for request/response mapping
   - **Add Swagger annotations to all handler functions**

4. **Wire Dependencies** (`internal/usecase/container.go`):
   - Add `Loan book.Repository` to `Repositories` struct
   - Add `CreateLoan *loanops.CreateLoanUseCase` to `Container` struct
   - Update `NewContainer()` to inject dependencies

5. **Migrations**:
   ```bash
   make migrate-create name=create_loans_table
   # Edit migrations/postgres/XXXXXX_create_loans_table.up.sql
   make migrate-up
   ```

6. **Update API Documentation**:
   ```bash
   # Regenerate Swagger docs after adding annotations
   make gen-docs
   # Verify at http://localhost:8080/swagger/index.html
   ```

## Key Implementation Patterns

### 1. Package Naming Convention

**Use Case Packages Use "ops" Suffix:**
- Domain: `internal/domain/book` (package `book`)
- Use Case: `internal/usecase/bookops` (package `bookops`)

**Rationale:**
- Avoids naming conflicts when importing both domain and use case packages
- No need for import aliases (cleaner, more idiomatic Go)
- Clear distinction: domain = entities/business rules, use cases = operations/orchestration

**Example:**
```go
import (
    "library-service/internal/domain/book"      // package book
    "library-service/internal/usecase/bookops"  // package bookops
)

// Clean references without aliases
bookEntity := book.NewEntity(...)
useCase := bookops.NewCreateBookUseCase(...)
```

### 2. Dependency Injection (Two-Step Wiring)

**Step 1: Application Bootstrap** (`internal/infrastructure/app/app.go`)

Boot order:
1. Logger initialization
2. Config loading
3. **Repositories** (DB layer):
   - `WithMemoryStore()` for tests/development
   - `WithPostgresStore(dsn)` for production (runs migrations automatically)
4. **Caches** (Redis/Memory)
5. **Auth Services** (JWT + Password)
6. **Use Cases Container** - wires everything together
7. **HTTP Server** - receives use cases

**Step 2: Use Case Container** (`internal/usecase/container.go`)

When adding new features:
1. Add repository interface to `Repositories` struct
2. Add cache interface to `Caches` struct (if needed)
3. Add use case to `Container` struct
4. Create **domain service** in `NewContainer()` - e.g., `book.NewService()`
5. Wire use case with dependencies in return statement

**Critical Distinction:**
- **Infrastructure Services** (JWT, Password): Created in `app.go`, passed to container
- **Domain Services** (Book, Member): Created in `container.go` `NewContainer()` function

### 3. Domain Services vs Use Cases

**Domain Service** (`internal/domain/book/service.go`):
- Pure business rules (ISBN validation, constraints)
- NO external dependencies (no database, HTTP, frameworks)
- Pure functions when possible
- 100% test coverage (easy to achieve)

**Use Case** (`internal/usecase/bookops/create_book.go`):
- Orchestrates domain entities and services
- Calls domain service for validation
- Persists to repository
- Updates cache
- Returns domain entities (not DTOs)

### 4. Repository Pattern

**Interface:** Defined in `internal/domain/{entity}/repository.go`
**Implementation:** In `internal/adapters/repository/{type}/{entity}.go`

```go
// Domain defines the contract
// internal/domain/book/repository.go
type Repository interface {
    Create(ctx context.Context, book Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    Update(ctx context.Context, book Entity) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ListFilter) ([]Entity, error)
}

// Adapter implements it
// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct {
    db *sqlx.DB
}

func (r *PostgresBookRepository) Create(ctx context.Context, book domain.Entity) error {
    query := `INSERT INTO books (id, name, isbn, genre) VALUES ($1, $2, $3, $4)`
    _, err := r.db.ExecContext(ctx, query, book.ID, book.Name, book.ISBN, book.Genre)
    return err
}
```

**Benefits:**
- Domain is independent of database technology
- Easy to swap PostgreSQL for MongoDB (just change adapter)
- Easy to mock for testing

### 5. Error Handling

```go
// Wrap errors with context (use %w for unwrapping)
if err := s.repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}

// Domain errors (defined in pkg/errors/domain.go)
return errors.ErrNotFound          // 404
return errors.ErrAlreadyExists     // 409
return errors.ErrValidation        // 400
```

## Authentication System

**JWT-based authentication with access/refresh tokens:**

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#","full_name":"John Doe"}'

# Login (returns access_token + refresh_token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#"}'

# Use access token for protected endpoints
curl -X GET http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer <access_token>"
```

**Token Configuration:**
- Access token: 24h (configurable via `JWT_EXPIRY`)
- Refresh token: 7 days
- Secret key: `JWT_SECRET` environment variable (MUST change in production)

**Protected Endpoints:**
All endpoints under `/api/v1/books/*` and `/api/v1/auth/me` require JWT authentication. These endpoints must have `@Security BearerAuth` in their Swagger annotations.

## Environment Configuration

**Setup:**
```bash
cp .env.example .env
# Edit .env with your settings (especially JWT_SECRET, DB credentials)
```

**Critical Variables:**
- `POSTGRES_DSN`: Database connection string
- `JWT_SECRET`: Token signing key (REQUIRED)
- `REDIS_HOST`: Cache server (optional, uses memory cache if unavailable)
- `APP_MODE`: `dev` (verbose logs) or `prod` (JSON logs)

**Docker Development:**
```bash
cd deployments/docker
docker-compose up -d  # PostgreSQL on :5432, Redis on :6379
```

## Testing Guidelines

**Unit Tests (Domain/Use Cases):**
```go
// Table-driven tests (Go standard)
func TestBookService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Integration Tests:**
- Use build tags: `//go:build integration`
- Test against real PostgreSQL (docker-compose)
- Run with: `make test-integration`

**Coverage Requirements:**
- Domain layer: 100% (critical business logic)
- Use cases: 80%+
- Overall: 60%+

## Dependency Management

```bash
go mod tidy             # Clean up dependencies
go mod vendor           # Vendor dependencies (project uses vendoring)
go get <package>        # Add new dependency
```

**Major Dependencies:**
- **Chi** (`go-chi/chi/v5`): HTTP router
- **sqlx** (`jmoiron/sqlx`): Database queries
- **Zap** (`uber.org/zap`): Structured logging
- **JWT** (`golang-jwt/jwt/v5`): Authentication
- **Validator** (`go-playground/validator/v10`): Input validation
- **Swaggo** (`swaggo/swag`, `swaggo/http-swagger`): API documentation

## Code Style Enforcement

**Linter Configuration:** `.golangci.yml` (25+ linters enabled)

**Key Rules:**
- Cyclomatic complexity: ≤10 per function
- Cognitive complexity: ≤20
- No naked returns in functions >30 lines
- All errors must be checked or explicitly ignored
- Context as first parameter in functions
- Errors as last return value

**Auto-fix:**
```bash
gofmt -w .              # Format code
goimports -w .          # Fix imports
```

## Troubleshooting

**"connection refused" errors:**
```bash
# Ensure PostgreSQL/Redis are running
make up
docker-compose -f deployments/docker/docker-compose.yml ps
```

**Migration errors:**
```bash
# Check database connection
psql -h localhost -U library -d library

# Reset database (destructive!)
make migrate-down
make migrate-up
```

**Swagger generation errors:**
```bash
# Ensure swag is installed
make install-tools

# Regenerate with full dependency parsing
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal

# Check for annotation errors in handler comments
```

**Build performance:**
- Build time: ~5 seconds (target)
- Test execution: ~2 seconds for unit tests
- Use `CGO_ENABLED=0` for static binaries (already in Makefile)

## Important Files

- `Makefile` - All common commands (30+ targets)
- `.golangci.yml` - Linter configuration
- `internal/usecase/container.go` - Dependency injection wiring
- `internal/infrastructure/app/app.go` - Application bootstrap
- `deployments/docker/docker-compose.yml` - Local development stack
- `migrations/postgres/` - Database schema changes
- `cmd/api/main.go` - API entry point and Swagger metadata
- `internal/adapters/http/router.go` - HTTP route configuration
- `api/openapi/` - Generated Swagger documentation

## Quick Reference

```bash
# Start coding (first time)
make init && make up && make migrate-up

# Daily development
make dev                # Start everything

# Before commit
make ci                 # Run full CI pipeline locally

# Add new feature (follow this order)
# 1. Domain (entity + service + tests)       → internal/domain/{entity}/
# 2. Use case (orchestration + tests)        → internal/usecase/{entity}ops/  (note "ops" suffix!)
# 3. Adapter (HTTP handler + repository)     → internal/adapters/
# 4. Add Swagger annotations to handlers     → @Security, @Summary, @Param, etc.
# 5. Wire in container.go                    → internal/usecase/container.go
# 6. Migration (if needed)                   → make migrate-create name=...
# 7. Regenerate API docs                     → make gen-docs
```

## Project-Specific Notes

### Current Use Cases Structure

**Book use cases** (`internal/usecase/bookops/`):
- CreateBook, GetBook, ListBooks, UpdateBook, DeleteBook, ListBookAuthors

**Auth use cases** (`internal/usecase/authops/`):
- RegisterMember, LoginMember, RefreshToken, ValidateToken

**Subscription use cases** (`internal/usecase/subops/`):
- SubscribeMember

### Domain Services

**Current domain services:**
- **BookService** (`internal/domain/book/service.go`): ISBN validation, business constraints
- **MemberService** (`internal/domain/member/service.go`): Subscription pricing logic

### Migration Locations
- **Postgres migrations:** `migrations/postgres/`
- **Naming:** Timestamped with descriptive names (e.g., `000001_create_books_table.up.sql`)
- **Always create both:** `.up.sql` and `.down.sql` files

### Test Data & Fixtures
- Shared test fixtures: `test/fixtures/`
- Integration test helpers: `test/testdb/setup.go`
- Build tags for integration tests: `//go:build integration`

### CI/CD Pipeline

**GitHub Actions Workflow** (`.github/workflows/ci.yml`):

The CI pipeline runs on push to `main`, `develop`, or `feature/*` branches and on PRs:

1. **Lint** - golangci-lint with project configuration
2. **Test** - All tests with coverage (PostgreSQL + Redis services)
3. **Build** - Multi-platform binaries (Linux, Darwin, Windows × amd64, arm64)
4. **Security** - gosec scanner + govulncheck for vulnerabilities
5. **Integration** - Integration tests (PR only)
6. **Docker** - Build Docker images (main/develop only)
7. **Quality** - SonarCloud scan + documentation checks

**Local CI Simulation:**
```bash
make ci  # Runs: fmt → vet → lint → test → build
```

**Key Requirements for PR:**
- All tests must pass
- Coverage maintained
- Linter passes
- No security vulnerabilities
- Integration tests pass
- Documentation updated

### Pre-approved Commands
These commands are safe to run without asking:
- `make test`, `make test-unit`, `make test-coverage`
- `make fmt`, `make vet`, `make lint`, `make ci`
- `go test ./internal/domain/...`
- `go run cmd/api/main.go` (local development)
- `make gen-docs` (regenerate Swagger documentation)
