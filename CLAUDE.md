# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚨 NEW CLAUDE CODE INSTANCE? **[START HERE → .claude/CLAUDE-START.md](./.claude/CLAUDE-START.md)**

> **📚 Full Documentation:** See [`.claude/`](./.claude/) directory for comprehensive guides
>
> **🎯 Not sure what to read?** Check [Context Guide](./.claude/context-guide.md) for task-specific reading lists

## Project Overview

Library Management System - A Go-based REST API following Clean Architecture principles, optimized for vibecoding with Claude Code. The system manages books, authors, members, and subscriptions with JWT authentication.

**Key Technologies:** Go 1.25, PostgreSQL, Redis, Chi router, JWT, Docker, Swagger/OpenAPI

## Documentation Index

### 🚀 Quick Start
- **[CLAUDE-START.md](./.claude/CLAUDE-START.md)** - 60-second boot sequence for new AI instances ⭐
- **[Context Guide](./.claude/context-guide.md)** - What to read for your specific task ⭐
- **[Quick Start & Navigation](./.claude/README.md)** - 30-second quick reference
- **[Cheat Sheet](./.claude/cheatsheet.md)** - Single-page command reference

### 📖 Essential Documentation
- **[Commands Reference](./.claude/commands.md)** - All essential commands
- **[Setup Guide](./.claude/setup.md)** - First-time setup and troubleshooting
- **[Architecture](./.claude/architecture.md)** - Clean architecture patterns
- **[Development Workflow](./.claude/development.md)** - Daily development tasks
- **[Testing Guide](./.claude/testing.md)** - Testing patterns and strategies
- **[API Documentation](./.claude/api.md)** - REST API endpoints
- **[Code Standards](./.claude/standards.md)** - Go best practices and conventions

### 🎯 Practical Guides
- **[Quick Wins](./.claude/quick-wins.md)** - Safe improvements you can suggest ⭐
- **[Development Workflows](./.claude/development-workflows.md)** - Complete workflows start to finish
- **[Debugging Guide](./.claude/debugging-guide.md)** - Advanced debugging techniques
- **[Gotchas](./.claude/gotchas.md)** - Common mistakes to avoid
- **[FAQ](./.claude/faq.md)** - Frequently asked questions
- **[Troubleshooting](./.claude/troubleshooting.md)** - Solutions to common problems

## Architecture (Clean Architecture)

The codebase follows strict dependency rules: **Domain → Use Case → Adapters → Infrastructure**

```
internal/
├── domain/              # Business logic (ZERO external dependencies)
│   ├── book/           # Book entity, service, repository interface
│   ├── member/         # Member entity, service (subscriptions)
│   ├── author/         # Author entity
│   └── reservation/    # Reservation entity, service (book reservations)
├── usecase/            # Application orchestration (depends on domain)
│   ├── bookops/        # CreateBook, UpdateBook, etc. ("ops" suffix)
│   ├── authops/        # Register, Login, RefreshToken ("ops" suffix)
│   ├── subops/         # SubscribeMember ("ops" suffix)
│   └── reservationops/ # CreateReservation, CancelReservation ("ops" suffix)
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
make gen-docs           # Generate Swagger/OpenAPI docs
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

**Important Swagger Annotations:**
- `@Summary` - Brief description (required)
- `@Description` - Detailed explanation
- `@Tags` - Group endpoints together
- `@Security BearerAuth` - **Required for all protected endpoints** (all /books routes, /auth/me)
- `@Param` - Request parameters (body, path, query, header)
- `@Success` / `@Failure` - Response codes with schemas
- `@Router` - Endpoint path and HTTP method

See [API Documentation Guide](./.claude/api.md) for complete details and examples.

## Development Workflow

### Adding a New Feature

**Follow this order:** Domain → Use Case → Adapters → Wiring → Migration → Documentation

See [Development Workflows](./.claude/development-workflows.md) for complete step-by-step guides.

**Quick Example: Adding a "Loan" domain**

1. **Domain Layer** (`internal/domain/loan/`):
   - Create entity, service, repository interface
   - Write unit tests with 100% coverage

2. **Use Case Layer** (`internal/usecase/loanops/`):
   - Note: "ops" suffix to avoid naming conflicts
   - Create use cases that orchestrate domain services
   - Mock repositories in tests

3. **Adapter Layer**:
   - Implement repository (`internal/adapters/repository/postgres/loan.go`)
   - Create HTTP handlers (`internal/adapters/http/handlers/loan.go`)
   - Add DTOs and Swagger annotations

4. **Wire Dependencies** (`internal/usecase/container.go`):
   - Add repository to `Repositories` struct
   - Add use cases to `Container` struct
   - Wire in `NewContainer()` function

5. **Database Migration**:
   ```bash
   make migrate-create name=create_loans_table
   make migrate-up
   ```

6. **Update Documentation**:
   ```bash
   make gen-docs
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
3. **Repositories** (DB layer)
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
All endpoints under `/api/v1/books/*`, `/api/v1/reservations/*`, and `/api/v1/auth/me` require JWT authentication.

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

See [Testing Guide](./.claude/testing.md) for comprehensive testing strategies.

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
make fmt                # Format code with gofmt + goimports
make vet                # Run go vet
make lint               # Run golangci-lint
```

## Troubleshooting

**Common Issues:**

**"connection refused" errors:**
```bash
make up
docker-compose -f deployments/docker/docker-compose.yml ps
```

**Migration errors:**
```bash
# Check database connection
psql -h localhost -U library -d library

# Reset database (destructive!)
make migrate-down && make migrate-up
```

**Port 8080 already in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**Tests fail randomly:**
```bash
go clean -testcache && make test
```

See [Troubleshooting Guide](./.claude/troubleshooting.md) for more solutions.

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

**Reservation use cases** (`internal/usecase/reservationops/`):
- CreateReservation, CancelReservation, GetReservation, ListMemberReservations

### Domain Services

**Current domain services:**
- **BookService** (`internal/domain/book/service.go`): ISBN validation, business constraints
- **MemberService** (`internal/domain/member/service.go`): Subscription pricing logic
- **ReservationService** (`internal/domain/reservation/service.go`): Reservation validation, status transitions, expiration logic

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
