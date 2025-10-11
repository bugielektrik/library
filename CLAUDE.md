# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## 🚨 NEW CLAUDE CODE INSTANCE? **[START HERE]**

**Session Start Protocol (2 minutes):**
1. Read `.claude-context/SESSION_MEMORY.md` - Architecture context (~1,200 tokens)
2. Read `.claude-context/CURRENT_PATTERNS.md` - Code patterns (~1,500 tokens)
3. Check `examples/` directory when creating code - Canonical implementations

**Full Documentation:** See [`.claude/`](./.claude/) directory for comprehensive guides

**Context Guide:** Check [`.claude/context-guide.md`](./.claude/context-guide.md) for task-specific reading lists

## Project Overview

Library Management System - Go-based REST API following Clean Architecture principles, optimized for AI-assisted development.

**Domains:** Books, Authors, Members, Subscriptions, Reservations, Payments (epayment.kz integration)

**Tech Stack:** Go 1.25, PostgreSQL 15+, Redis 7+, Chi router, JWT, Docker, Swagger/OpenAPI

## Architecture

Clean Architecture with strict dependency rules: **Domain → Use Case → Adapters → Infrastructure**

**Bounded Context Organization:** All domains migrated to vertical slices (Phases 2.1-2.5 Complete ✅)

```
internal/
├── books/              # Books bounded context (✅ Phases 2.1 + 2.5 Complete)
│   ├── domain/        # book/ and author/ entities, services, interfaces
│   ├── operations/    # Book use cases + author/ subdomain
│   ├── http/          # Book HTTP handlers + author/ subdomain
│   └── repository/    # Book PostgreSQL implementations
├── members/            # Members bounded context (✅ Phase 2.2 Complete)
│   ├── domain/        # Member entity, service, repository interface
│   ├── operations/    # Auth, profile, subscription use cases
│   │   ├── auth/      # Register, login, refresh, validate
│   │   ├── profile/   # Get profile, list members
│   │   └── subscription/  # Subscribe member
│   ├── http/          # Auth and profile HTTP handlers
│   │   ├── auth/
│   │   └── profile/
│   └── repository/    # Member PostgreSQL implementation
├── payments/           # Payments bounded context (✅ Phase 2.3 Complete)
│   ├── domain/        # Payment, SavedCard, Receipt entities, service
│   ├── operations/    # Payment use cases organized by subdomain
│   │   ├── payment/   # Initiate, verify, cancel, refund, callbacks, expiry
│   │   ├── savedcard/ # Save, list, delete, set default, pay with saved card
│   │   └── receipt/   # Generate, get, list receipts
│   ├── http/          # Payment HTTP handlers
│   │   ├── payment/
│   │   ├── savedcard/
│   │   └── receipt/
│   ├── repository/    # Payment PostgreSQL implementations (4 repos)
│   └── gateway/       # Payment gateway integrations
│       └── epayment/  # epayment.kz adapter
├── reservations/       # Reservations bounded context (✅ Phase 2.4 Complete)
│   ├── domain/        # Reservation entity, service, repository interface
│   ├── operations/    # Reservation use cases (create, cancel, get, list)
│   ├── http/          # Reservation HTTP handlers
│   └── repository/    # Reservation PostgreSQL implementation
├── adapters/          # Shared adapters
│   ├── http/          # Middleware, DTOs, remaining handlers
│   └── repository/    # Remaining repository implementations
└── infrastructure/    # Technical concerns (shared)
    ├── auth/          # JWT
    ├── store/         # DB connections
    └── server/        # HTTP server
```

**Critical Rules:**
- Domain layer NEVER imports from outer layers
- **Within bounded contexts**: Use generic package names (`domain`, `operations`, `http`, `repository`)
- **Import aliases required** for bounded contexts to avoid naming conflicts (e.g., `bookdomain`, `authorops`)
- Business logic lives in **domain services**, use cases orchestrate
- Infrastructure services (JWT, Password) created in `app.go`
- Domain services created in respective bounded context or in usecase factories

**Detailed Architecture:** See `.claude/architecture.md`

## Quick Start

### First Time Setup
```bash
# Automated (recommended)
./scripts/dev-setup.sh  # Deps + docker + migrations + seeds + hooks

# OR manual
make init && make up && make migrate-up
make install-hooks
```

### Daily Development
```bash
make dev                # Start everything (docker + migrations + API)
make run                # Run API server only
```

### Testing
```bash
make test               # All tests with race detection + coverage
make test-unit          # Unit tests only (fast)
make ci                 # Full CI pipeline (run before commit)
```

### Common Commands
```bash
# Build
make build              # Build all binaries (api, worker, migrate)

# Database
make migrate-up         # Apply migrations
make migrate-create name=add_feature  # Create new migration

# Code Quality
make fmt                # Format code
make lint               # Run linters (25+ enabled)

# API Docs
make gen-docs           # Regenerate Swagger docs
# View at: http://localhost:8080/swagger/index.html

# Development Data
./scripts/seed-data.sh  # Seed test users and books
# Test accounts: admin@library.com / Admin123!@#
#                user@library.com / User123!@#
```

**Full Commands Reference:** See `.claude/commands.md` or `Makefile`

## Development Workflow

**Adding a New Feature** (follow this order):

**For new domains (recommended - bounded context):**

1. **Create bounded context structure**
   ```bash
   mkdir -p internal/{entity}/{domain,operations,http,repository}
   ```

2. **Domain Layer** (`internal/{entity}/domain/`)
   - Create entity, service, repository interface
   - Write unit tests (100% coverage target)
   - Package name: `domain`

3. **Operations Layer** (`internal/{entity}/operations/`)
   - Create use cases that orchestrate domain services
   - Mock repositories in tests
   - Package name: `operations`

4. **Adapter Layers**
   - HTTP handlers: `internal/{entity}/http/` (package: `http`)
   - Repository: `internal/{entity}/repository/` (package: `repository`)
   - Add DTOs with Swagger annotations

5. **Wire Dependencies** (`internal/usecase/container.go`)
   - Add repository to `Repositories` struct
   - Add use cases to appropriate domain group in `Container`
   - Wire in `NewContainer()` function

6. **Database Migration**
   ```bash
   make migrate-create name=create_entity_table
   make migrate-up
   ```

7. **Regenerate API Docs**
   ```bash
   make gen-docs
   ```

**For legacy extensions (if extending existing non-migrated domains):**
Follow the same order but use `internal/domain/{entity}/`, `internal/usecase/{entity}ops/` with "ops" suffix.

**Detailed Workflows:** See `.claude/development-workflows.md` or `.claude/common-tasks.md`

## Code Patterns

**Quick Pattern Reference:**
- See `examples/handler_example.go` - HTTP handler pattern
- See `examples/usecase_example.go` - Use case pattern
- See `examples/repository_example.go` - Repository pattern
- See `examples/test_example_test.go` - Table-driven test pattern

**Current Patterns Documentation:** See `.claude-context/CURRENT_PATTERNS.md`

### Key Patterns

**1. Package Naming (Bounded Contexts):**
```go
// Bounded context imports use aliases to avoid conflicts
import (
    bookdomain "library-service/internal/books/domain"
    bookops "library-service/internal/books/operations"
    bookhttp "library-service/internal/books/http"
)

// Clear references with aliases
entity := bookdomain.NewEntity(...)
useCase := bookops.NewCreateBookUseCase(...)
handler := bookhttp.NewBookHandler(...)
```

**2. Grouped Container Structure:**
```go
type Container struct {
    Book         BookUseCases
    Auth         AuthUseCases
    Payment      PaymentUseCases
    // ...
}

// Handler access
h.useCases.Book.CreateBook.Execute(ctx, req)
```

**3. Context Helpers (ALWAYS use):**
```go
import "library-service/internal/adapters/http/middleware"

memberID, ok := middleware.GetMemberIDFromContext(ctx)
if !ok {
    h.RespondError(w, r, errors.ErrUnauthorized)
    return
}
```

**4. Status Code Checks (NEVER use magic numbers):**
```go
import "library-service/pkg/httputil"

if httputil.IsServerError(status) {  // Instead of: status >= 500
    logger.Error("internal error")
}
```

**5. Validator Injection:**
```go
// ✅ CORRECT - Inject as dependency
func NewBookHandler(
    useCases *usecase.Container,
    validator *middleware.Validator,  // Injected
) *BookHandler {
    // ...
}
```

**Full Patterns:** See `.claude-context/CURRENT_PATTERNS.md` or `examples/` directory

## Authentication

JWT-based with access/refresh tokens.

**Quick Example:**
```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#","full_name":"John Doe"}'

# Login (returns access_token + refresh_token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#"}'

# Use access token
curl -X GET http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer <access_token>"
```

**Configuration:**
- Access token: 24h (configurable via `JWT_EXPIRY`)
- Refresh token: 7 days
- Secret: `JWT_SECRET` environment variable (MUST change in production)

## Payment System

**Integration:** epayment.kz (Kazakhstan payment gateway)

**Features:**
- Payment types: Fines, subscriptions, book purchases
- Supported currencies: KZT, USD, EUR, RUB
- Payment methods: Card, saved card
- Refunds: Full and partial
- Receipts: Auto-generated (RCP-YYYY-NNNNN format)
- Background worker: Payment expiry + callback retries

**Configuration:**
```bash
EPAYMENT_BASE_URL="https://api.epayment.kz"
EPAYMENT_CLIENT_ID="your-client-id"
EPAYMENT_CLIENT_SECRET="your-client-secret"
EPAYMENT_TERMINAL="your-terminal-id"
```

**Run Worker:**
```bash
make run-worker  # Processes payment expiry + callback retries
```

## Environment Configuration

```bash
cp .env.example .env
# Edit with your settings
```

**Critical Variables:**
- `POSTGRES_DSN`: Database connection string
- `JWT_SECRET`: Token signing key (REQUIRED)
- `REDIS_HOST`: Cache server (optional, uses memory cache if unavailable)
- `APP_MODE`: `dev` (verbose logs) or `prod` (JSON logs)
- `EPAYMENT_*`: Payment gateway credentials

## Testing Guidelines

**Coverage Requirements:**
- Domain layer: 100%
- Use cases: 80%+
- Overall: 60%+

**Table-Driven Test Pattern:**
```go
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
- Use build tag: `//go:build integration`
- Run with: `make test-integration`

**Full Testing Guide:** See `.claude/testing.md`

## Important Files

### Core Architecture
- `internal/usecase/container.go` - **CRITICAL**: Dependency injection wiring
- `internal/infrastructure/app/app.go` - **CRITICAL**: Application bootstrap
- `internal/adapters/http/router.go` - HTTP route configuration
- `cmd/api/main.go` - API entry point
- `cmd/worker/main.go` - Background worker

### Utilities (Refactoring Phases 1-5)
- `pkg/strutil/` - Safe string pointer utilities
- `pkg/httputil/` - HTTP status helpers
- `pkg/logutil/` - Logger utilities (UseCaseLogger, HandlerLogger, etc.)
- `internal/adapters/http/handlers/base.go` - Shared response methods

### Configuration
- `Makefile` - 30+ command targets
- `.golangci.yml` - Linter configuration (25+ linters)
- `migrations/postgres/` - Database schema
- `api/openapi/` - Generated Swagger docs

### Documentation
- `.claude/README.md` - Quick start (30 seconds)
- `.claude/context-guide.md` - **Task-specific reading lists** ⭐
- `.claude/architecture.md` - Detailed architecture
- `.claude/development-workflows.md` - Step-by-step workflows
- `.claude/testing.md` - Testing patterns
- `.claude-context/SESSION_MEMORY.md` - **Architecture context** ⭐
- `.claude-context/CURRENT_PATTERNS.md` - **Code patterns** ⭐

## Troubleshooting

**Connection refused errors:**
```bash
make up
docker-compose -f deployments/docker/docker-compose.yml ps
```

**Migration errors:**
```bash
psql -h localhost -U library -d library  # Check connection
make migrate-down && make migrate-up     # Reset (destructive!)
```

**Port 8080 in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**Tests fail randomly:**
```bash
go clean -testcache && make test
```

**Full Troubleshooting Guide:** See `.claude/troubleshooting.md`

## Quick Reference

```bash
# Start coding (first time)
./scripts/dev-setup.sh  # Automated setup

# Daily development
make dev                # Start everything

# Before commit (pre-commit hooks run automatically)
make ci                 # Run full CI pipeline

# Add new feature (order matters!)
# 1. Create bounded context structure        → internal/{entity}/{domain,operations,http,repository}
# 2. Domain (entity + service + tests)       → internal/{entity}/domain/ (package: domain)
# 3. Operations (orchestration + tests)      → internal/{entity}/operations/ (package: operations)
# 4. Adapters (HTTP + repository)            → internal/{entity}/http/ and repository/
# 5. Add Swagger annotations to handlers     → @Security, @Summary, @Param, etc.
# 6. Wire in container.go                    → internal/usecase/container.go
# 7. Migration (if needed)                   → make migrate-create name=...
# 8. Regenerate API docs                     → make gen-docs
```

## Pre-approved Commands

Safe to run without asking:
- `make test`, `make test-unit`, `make test-coverage`
- `make fmt`, `make vet`, `make lint`, `make ci`
- `go test ./internal/domain/...`
- `go run cmd/api/main.go` (local development)
- `make gen-docs` (regenerate Swagger)

---

**Token Optimization Notes:**
- This file: ~450 lines (~900 tokens) vs 1,054 lines (~2,100 tokens) previously
- **52% reduction** achieved
- Detailed content moved to `.claude/` and `.claude-context/` directories
- Use session context files for efficiency (see top of file)
