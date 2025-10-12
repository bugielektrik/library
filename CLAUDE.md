# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚨 NEW CLAUDE CODE INSTANCE? **[START HERE]**

**Session Start Protocol (8 minutes):**
1. Read this file (CLAUDE.md) - 2 minutes
2. Read `.claude-context/SESSION_MEMORY.md` - Architecture context - 3 minutes
3. Read `.claude-context/CURRENT_PATTERNS.md` - Code patterns - 3 minutes
4. Check `examples/` directory when creating code - Canonical implementations

**Full Documentation:** See [`.claude/`](./.claude/) directory for comprehensive guides

**Navigation:** Check [`.claude/README.md`](./.claude/README.md) for organized documentation index

## Project Overview

Library Management System - Go-based REST API following Clean Architecture principles, optimized for AI-assisted development.

**Domains:** Books, Authors, Members, Subscriptions, Reservations, Payments (epayment.kz integration)

**Tech Stack:** Go 1.25, PostgreSQL 15+, Redis 7+, Chi router, JWT, Docker, Swagger/OpenAPI

## Architecture

Clean Architecture with strict dependency rules: **Domain → Use Case → Adapters → Infrastructure**

**Bounded Context Organization:** All domains migrated to vertical slices (Phases 2.1-2.5 Complete ✅)

**Current Structure** (as of October 2025 - vibecoding branch):

```
internal/
├── books/              # Books bounded context (✅ Complete)
│   ├── domain/        # book/ and author/ entities, services, interfaces
│   ├── service/    # Book use cases + author/ subdomain
│   ├── handler/       # Book HTTP handlers + author/ subdomain + DTOs
│   ├── repository/    # Book PostgreSQL implementations + memory (tests)
│   │   └── mocks/     # Auto-generated test mocks (✅ Phase 2.1)
│   └── cache/         # ✅ Cache implementations (memory, redis)
├── members/            # Members bounded context (✅ Phase 2.2 Complete)
│   ├── domain/        # Member entity, service, repository interface
│   ├── service/    # Auth, profile, subscription use cases
│   │   ├── auth/      # Register, login, refresh, validate
│   │   ├── profile/   # Get profile, list members
│   │   └── subscription/  # Subscribe member
│   ├── handler/       # Auth and profile HTTP handlers + DTOs
│   │   ├── auth/
│   │   └── profile/
│   └── repository/    # Member PostgreSQL implementation + memory (tests)
│       └── mocks/     # Auto-generated test mocks (✅ Phase 2.1)
├── payments/           # Payments bounded context (✅ Token-Optimized Oct 2025)
│   ├── domain/        # Payment, SavedCard, Receipt entities, service
│   ├── service/    # Consolidated service files (20 → 4 files, 50-60% token reduction)
│   │   ├── payment/   # payment_operations.go (818 lines), payment_callbacks.go (449 lines)
│   │   ├── savedcard/ # saved_card.go (342 lines - delete, list, pay with card)
│   │   └── receipt/   # receipt.go (284 lines - generate, get, list)
│   ├── handler/       # HTTP handlers (token-efficient organization)
│   │   ├── payment/   # 5 handler files (73-129 lines each) + 3 DTO files (split by feature)
│   │   ├── savedcard/ # handler.go (270 lines consolidated) + dto.go
│   │   └── receipt/   # handler.go (192 lines) + dto.go (101 lines)
│   ├── repository/    # Payment PostgreSQL implementations
│   │   └── postgres/  # 4 repositories (payment, receipt, saved_card, callback_retry)
│   └── provider/      # Payment gateway integrations
│       └── epayment/  # epayment.kz adapter
├── reservations/       # Reservations bounded context (✅ Phase 2.4 Complete)
│   ├── domain/        # Reservation entity, service, repository interface
│   ├── service/    # Reservation use cases (create, cancel, get, list)
│   ├── handler/       # Reservation HTTP handlers + DTOs
│   └── repository/    # Reservation PostgreSQL implementation
│       └── mocks/     # Auto-generated test mocks (✅ Phase 2.1)
├── container/          # ✅ Use case container (consolidated)
│   └── container.go   # All use case wiring & dependency injection
├── app/               # ✅ Application layer (domain-aware wiring)
│   ├── app.go         # Application bootstrap
│   ├── repository.go  # Repository container (domain-aware)
│   ├── cache.go       # Cache container (domain-aware)
│   ├── warming.go     # Cache warming logic
│   └── warming_test.go
├── pkg/               # ✅ Shared utilities (domain-agnostic)
│   ├── repository/    # Postgres utilities only (BaseRepository, helpers)
│   │   └── postgres/  # Generic SQL utilities (shared by all repos)
│   ├── handlers/      # Base handler utilities
│   ├── middleware/    # Auth, error, logging, validation middleware
│   ├── errors/        # Domain-specific error types
│   ├── httputil/      # HTTP helpers (status, JSON)
│   ├── logutil/       # Logger factories
│   ├── pagination/    # Pagination helpers
│   ├── sqlutil/       # SQL null conversion
│   └── strutil/       # String pointer utilities
└── infrastructure/    # Technical concerns (shared) ✅ Clean Architecture compliant
    ├── auth/          # JWT, Password services
    ├── config/        # Viper wrapper with validation
    ├── log/           # Logging configuration
    ├── store/         # DB connections
    ├── server/        # ✅ HTTP server (moved to break import cycles)
    │   ├── http.go    # Server initialization
    │   └── router.go  # Route configuration
    └── shutdown/      # Graceful shutdown manager
```

**Critical Rules:**
- **Domain layer NEVER imports from outer layers** (Clean Architecture)
- **Infrastructure layer NEVER imports domain/use cases/adapters** (Critical!)
  - Infrastructure must be generic and reusable
  - Accept primitive types (string, int, bool), not domain types
  - Example: `JWTService.GenerateAccessToken(..., role string)` not `role domain.Role`
  - Type conversion happens at use case layer: `string(domain.RoleUser)`
  - See `CLEAN_ARCHITECTURE_FIX.md` for detailed explanation
- **Within bounded contexts**: Use generic package names (`domain`, `service`, `handler`, `repository`)
- **Import aliases required** for cross-context imports (✅ Phase 2.2 Standardized):
  - Pattern: `{context}domain`, `{context}service`, `{context}handler`, `{context}repo`
  - Examples: `bookdomain`, `memberservice`, `paymenthandler`, `reservationrepo`
  - All directories are now `handler/` (not `http/` or `handlers/`)
- **DTOs colocated** with HTTP handlers (✅ Phase 1.1 Payment DTOs Split):
  - Large subdomain handler packages: Each subdomain has its own `dto.go`
  - Example: `internal/payments/handler/payment/dto_core.go`, `dto_operations.go`, `dto_callback.go`
- **Auto-generated mocks** in bounded contexts (✅ Phase 2.1):
  - Location: `internal/{context}/repository/mocks/`
  - Configured via `.mockery.yaml` (30+ interfaces)
  - Generate: `make gen-mocks` (uses mockery tool)
- **Memory repositories** for testing (✅ All domains: books, members, payments, reservations):
  - Location: `internal/{context}/repository/memory/`
  - Thread-safe with sync.RWMutex
  - Full interface implementation with compile-time verification
- **Cache implementations** colocated in `cache/` within bounded contexts (e.g., `books/cache/`)
- **Compile-time interface verification**: All repositories use `var _ Interface = (*Implementation)(nil)`
- Business logic lives in **domain services**, use cases orchestrate
- Infrastructure services (JWT, Password) created in `app.go`
- Domain services created in respective bounded context or in usecase factories

**Detailed Architecture:** See `.claude/guides/architecture.md`

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

# Run specific package tests
go test ./internal/books/service/... -v
go test ./internal/members/service/auth/... -run TestLogin

# Run with coverage
go test ./internal/books/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out  # View in browser
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

**Full Commands Reference:** See `.claude/guides/development.md` or `Makefile`

## Development Workflow

**Adding a New Feature** (follow this order):

**For new domains (required - bounded context):**

1. **Create bounded context structure**
   ```bash
   mkdir -p internal/{entity}/{domain,service,handler,repository}
   ```

2. **Domain Layer** (`internal/{entity}/domain/`)
   - Create entity, service, repository interface
   - Write unit tests (100% coverage target)
   - Package name: `domain`

3. **Service Layer** (`internal/{entity}/service/`)
   - Create use cases that orchestrate domain services
   - Mock repositories in tests
   - Package name: `service`

4. **Adapter Layers**
   - HTTP handlers: `internal/{entity}/handler/` (package: `handler` or subdirectory name)
   - DTOs: `internal/{entity}/handler/dto.go` (colocated with handlers)
     - For large subdomains: Split into subdomain-specific DTOs (see payments example)
   - Repository: `internal/{entity}/repository/` (package: `repository`)
   - Memory repo (tests): `internal/{entity}/repository/memory/` (optional)
   - Mocks: `internal/{entity}/repository/mocks/` (auto-generated via `make gen-mocks`)
   - Add Swagger annotations to handlers

5. **Wire Dependencies** (`internal/container/container.go`)
   - Add repository to `Repositories` struct
   - Add use cases to appropriate domain group in `Container`
   - Wire in `NewContainer()` function

6. **Database Migration**
   ```bash
   make migrate-create name=create_entity_table
   make migrate-up
   ```

7. **Generate Mocks** (if adding repository interfaces)
   ```bash
   # Update .mockery.yaml with new interface
   make gen-mocks
   ```

8. **Regenerate API Docs**
   ```bash
   make gen-docs
   ```

**Note:** All domains now use bounded context structure. Legacy layered structure has been fully migrated.

**Detailed Workflows:** See `.claude/guides/common-tasks.md`

## Code Patterns

**Quick Pattern Reference:**
- See `examples/handler_pattern.md` - HTTP handler pattern
- See `examples/usecase_pattern.md` - Use case pattern
- See `examples/repository_pattern.md` - Repository pattern
- See `examples/testing_pattern.md` - Testing patterns

**Current Patterns Documentation:** See `.claude-context/CURRENT_PATTERNS.md`

### Key Patterns

**1. Package Naming & Import Aliases (✅ Phase 2.2 Standardized):**
```go
// Cross-context imports MUST use standardized aliases
import (
    // Domain imports - use {context}domain
    bookdomain "library-service/internal/books/domain/book"
    memberdomain "library-service/internal/members/domain"

    // Service imports - use {context}service
    bookservice "library-service/internal/books/service"
    paymentservice "library-service/internal/payments/service/payment"

    // Handler imports - use {context}handler
    bookhandlers "library-service/internal/books/handler"
    paymenthandlers "library-service/internal/payments/handler/payment"

    // Repository imports - use {context}repo
    bookrepo "library-service/internal/books/repository"

    // Mock imports - use {context}mocks (✅ Phase 2.1 - auto-generated in bounded contexts)
    bookmocks "library-service/internal/books/repository/mocks"
    membermocks "library-service/internal/members/repository/mocks"
)

// DTOs colocated with handler (✅ Phase 1.1 - split for large subdomains)
// Payment subdomains each have their own dto.go:
var req paymenthandler.InitiatePaymentRequest  // from payment/dto.go
var cardReq savedcardhandler.SaveCardRequest   // from savedcard/dto.go
var receiptReq receipthandler.GenerateReceiptRequest  // from receipt/dto.go

// Usage examples
entity := bookdomain.Book{...}
useCase := bookservice.NewCreateBookUseCase(...)
handler := bookhandler.NewBookHandler(...)
mock := bookmocks.NewMockBookRepository(t)
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
import "library-service/internal/pkg/middleware"

memberID, ok := middleware.GetMemberIDFromContext(ctx)
if !ok {
    h.RespondError(w, r, errors.ErrUnauthorized)
    return
}
```

**Note:** Middleware was moved from `internal/infrastructure/pkg/middleware` to `internal/pkg/middleware` in Phase 5 (October 2025)

**4. Status Code Checks (NEVER use magic numbers):**
```go
import "library-service/internal/pkg/httputil"

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

## Cache System

**Implementation:** Redis (production) or Memory (development/testing)

**Cache Warming:** Automatic pre-loading on startup
- Warms top 50 books and 20 authors by default
- Runs asynchronously (non-blocking startup)
- Configurable limits and 30s timeout
- Reduces latency for frequently accessed items

**Configuration:**
```go
// Defaults in internal/app/warming.go
app.DefaultWarmingConfig(logger) // Auto-called on startup
```

**Documentation:** See `.claude/guides/cache-warming.md` for details

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
make run-worker  # Background worker for payment processing
```

**Worker Functions:**
- Expires pending payments (15-minute timeout)
- Retries failed payment callbacks (exponential backoff)
- Runs continuously; stop with Ctrl+C

## Environment Configuration

**Configuration System:** Viper (supports .env files, environment variables, YAML/JSON/TOML)

```bash
cp .env.example .env
# Edit with your settings
```

**Critical Variables:**
- `POSTGRES_DSN`: Database connection string
- `JWT_SECRET`: Token signing key (REQUIRED, minimum 32 characters)
- `JWT_EXPIRY`: Access token TTL (default: 24h)
- `REDIS_HOST`: Cache server (optional, uses memory cache if unavailable)
- `APP_MODE`: `dev` (verbose logs) or `prod` (JSON logs)
- `EPAYMENT_*`: Payment gateway credentials

**Configuration Loading:**
- Automatic environment variable binding with `VIPER_` prefix support
- Hot-reload capability for config files
- Validation on startup
- Bridge layer for backward compatibility (internal/infrastructure/config/)

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

**Full Testing Guide:** See `.claude/guides/testing.md`

## Important Files

### Core Architecture
- `internal/container/container.go` - **CRITICAL**: Use case dependency injection
- `internal/app/app.go` - **CRITICAL**: Application bootstrap
- `internal/app/repository.go` - **CRITICAL**: Repository container (domain-aware)
- `internal/app/cache.go` - **CRITICAL**: Cache container (domain-aware)
- `internal/app/warming.go` - Cache warming logic
- `internal/infrastructure/server/router.go` - HTTP route configuration
- `cmd/api/main.go` - API entry point
- `cmd/worker/main.go` - Background worker

### Utilities (Refactoring Phases 1-5)
- `internal/pkg/strutil/` - Safe string pointer utilities
- `internal/pkg/httputil/` - HTTP status helpers
- `internal/pkg/logutil/` - Logger utilities (UseCaseLogger, HandlerLogger, etc.)
- `internal/pkg/errors/` - 35 domain-specific sentinel errors across 6 domains
- `internal/infrastructure/config/` - Viper-based configuration loader (replaced godotenv+envconfig)
- `internal/pkg/pagination/` - Cursor and offset pagination helpers
- `internal/pkg/sqlutil/` - SQL null type conversion helpers (reduces repository boilerplate)
- `internal/pkg/handlers/base.go` - Shared response methods

### Configuration & Infrastructure
- `Makefile` - 30+ command targets
- `.golangci.yml` - Linter configuration (25+ linters)
- `migrations/postgres/` - Database schema
- `api/openapi/` - Generated Swagger docs
- `internal/app/repository.go` - Repository container (WithMemoryStore, WithPostgresStore)
- `internal/app/cache.go` - Cache container (WithMemoryCache, WithRedisCache)
- `internal/pkg/repository/postgres/` - Shared PostgreSQL utilities (BaseRepository, helpers with generic PrepareUpdateArgs)
- `internal/infrastructure/shutdown/` - Phased graceful shutdown manager (5 phases, hook system)
- `internal/pkg/middleware/` - Custom middleware (auth, error, request_logger, validator)
  - Chi built-ins used: RequestID, Recoverer, RealIP, Timeout, Heartbeat
  - Custom: Auth (JWT), Error handler, Request logger, Validator

### Documentation
- `.claude/README.md` - **Documentation hub** ⭐
- `.claude/guides/architecture.md` - Detailed architecture
- `.claude/guides/development.md` - Development guide
- `.claude/guides/common-tasks.md` - Step-by-step workflows
- `.claude/guides/testing.md` - Testing patterns
- `.claude/guides/cache-warming.md` - Cache warming implementation
- `.claude/reference/common-mistakes.md` - Gotchas to avoid
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

**Config validation errors:**
```bash
# Check JWT_SECRET is set and >= 32 characters
echo $JWT_SECRET | wc -c

# View loaded config
go run cmd/api/main.go --print-config  # If supported
```

**Graceful shutdown issues:**
- Check logs for shutdown phase timeouts (default: 30s total)
- Phases: pre_shutdown (2s), stop_accepting (1s), drain (10s), cleanup (5s), post_shutdown (2s)
- Increase timeout in `internal/app/app.go` if needed

**Full Troubleshooting Guide:** See `.claude/guides/development.md`

## Quick Reference

```bash
# Start coding (first time)
./scripts/dev-setup.sh  # Automated setup

# Daily development
make dev                # Start everything

# Before commit (pre-commit hooks run automatically)
make ci                 # Run full CI pipeline

# Add new feature (order matters!)
# 1. Create bounded context structure        → internal/{entity}/{domain,service,handler,repository}
# 2. Domain (entity + service + tests)       → internal/{entity}/domain/ (package: domain)
# 3. Service (orchestration + tests)         → internal/{entity}/service/ (package: service)
# 4. Adapters (HTTP + repository)            → internal/{entity}/handler/ and repository/
# 5. Add Swagger annotations to handler     → @Security, @Summary, @Param, etc.
# 6. Wire in container.go                    → internal/container/container.go
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

## Repository & Cache Organization

**Application Layer** (`internal/app/` - domain-aware wiring):
- `repository.go` - Repository container with `WithMemoryStore()` and `WithPostgresStore()`
- `cache.go` - Cache container with `WithMemoryCache()` and `WithRedisCache()`
- `warming.go` - Cache warming logic

**Shared Utilities Layer** (`internal/pkg/repository/` - domain-agnostic utilities):
- `postgres/` - BaseRepository pattern, SQL helpers (shared by all bounded contexts)
  - `base.go` - Generic CRUD operations with Go generics
  - `helpers.go` - SQL error handling, PrepareUpdateArgs
  - `generic.go` - Generic query builders

**Bounded Context Specific** (in each domain):
- `internal/{entity}/repository/` - PostgreSQL implementations
- `internal/{entity}/repository/memory/` - In-memory implementations for unit tests

**Example Usage:**
```go
// Application bootstrap (internal/app/app.go)
import "library-service/internal/app"

repos, err := app.NewRepositories(app.WithMemoryStore())  // For tests
repos, err := app.NewRepositories(app.WithPostgresStore(dsn))  // For production

caches, err := app.NewCaches(
    app.Dependencies{Repositories: repos},
    app.WithMemoryCache(),  // For dev/test
)

// Bounded context repos use shared postgres utilities
import "library-service/internal/pkg/repository/postgres"

type BookRepository struct {
    postgres.BaseRepository[book.Book]  // Inherits List, Get, Delete, etc.
}

// All repositories have compile-time interface verification
var _ domain.Repository = (*BookRepository)(nil)
```

---

## 📚 Additional Resources

**Documentation refactored October 2025:**
- **60% reduction** in documentation files (77 → 31 active)
- Clear organization: `guides/`, `adr/`, `reference/`, `archive/`
- Payment docs organized in `docs/payments/`
- Historical docs preserved in archives

**For comprehensive documentation index:**
→ See [`.claude/README.md`](./.claude/README.md) for full navigation

---

**Token Optimization Notes:**
- This file: ~560 lines (~1,120 tokens) - Core entry point
- Session context files: ~2,700 tokens total (SESSION_MEMORY + CURRENT_PATTERNS)
- Full documentation: 31 active files (vs 77 before refactoring)
- **Total onboarding: 8 minutes** to full productivity

---

## 🚀 Production Readiness (October 2025 Refactoring)

**Infrastructure Improvements:**
- ✅ **Phased Graceful Shutdown** - 5-phase shutdown with hook system (internal/infrastructure/shutdown/)
- ✅ **Viper Configuration** - Industry-standard config management with hot-reload (internal/infrastructure/config/)
- ✅ **Package Organization** - All utilities moved to internal/pkg/ (Phase 5 - Oct 2025)
- ✅ **Domain-Specific Errors** - 35 sentinel errors across 6 domains for precise error handling
- ✅ **Complete Test Infrastructure** - All 4 bounded contexts have memory repositories
- ✅ **Compile-Time Safety** - All 8 PostgreSQL repositories verified at compile time
- ✅ **Generic Helpers** - PrepareUpdateArgs with automatic PostgreSQL array handling
- ✅ **Industry Standards** - Chi, Zap, Viper, Testify, JWT/v5, Decimal, sqlx, go-playground/validator

**Code Quality Metrics:**
- 100% compile-time interface verification on repositories
- Thread-safe memory repositories with RWMutex
- Zero deprecated functions
- Consistent context propagation across all repository methods
- DTO files organized by subdomain (payment split: 626 lines → 3 files × ~200 lines)
- Chi built-in middleware (no custom duplicates)

**Recent Refactoring (October 2025):**

**Phase 4A - Cleanup:**
- ✅ Removed 536 lines of unused/duplicate code
- ✅ Removed 4 unused pkg packages (validator, crypto, timeutil, constants)
- ✅ Removed unused infrastructure/server abstraction (137 lines)
- ✅ Removed duplicate middleware (Chi provides RequestID, Recoverer)
- ✅ Cleaned up committed log files
- ✅ All tests passing with maintained coverage

**Phase 4B - Configuration:**
- ✅ Cleaned .env.example (removed 35 lines of non-existent features)
- ✅ Removed log files from source control
- ✅ APM analysis documented
- ✅ Zero breaking changes

**Phase 5 - Package Organization (October 2025):**
- ✅ Utilities organized in `internal/pkg/` (shared, domain-agnostic)
- ✅ Infrastructure services in `internal/infrastructure/` (auth, config, log, store, shutdown)
- ✅ Clear separation: utilities vs infrastructure concerns
- ✅ All tests passing, all builds successful

**Clean Architecture Compliance (October 2025):**
- ✅ Fixed infrastructure layer to be domain-agnostic
- ✅ Moved domain-aware containers to `internal/app/` layer (repository.go, cache.go, warming.go)
- ✅ Auth middleware uses strings instead of domain types (e.g., `RequireRole("admin")`)
- ✅ Shared utilities in `internal/pkg/` (domain-agnostic)
- ✅ Type conversion happens at handler layer (correct location in clean architecture)
- ✅ Zero domain dependencies in `internal/infrastructure/`
- ✅ All tests passing (60+ tests), all builds successful

**Phase 7 - Repository Optimization (October 2025):**
- ✅ Removed vendor directory (2,568 files, -58MB)
- ✅ Repository size reduced: 249MB → 70MB (72% reduction!)
- ✅ Removed empty `internal/adapters/` directory structure (leftover from Phase 6)
- ✅ Removed 3 backup test files (.backup, .backup2)
- ✅ Removed 4 empty app directories (bounded context artifacts)
- ✅ Added `/vendor/` to `.gitignore` for modern Go module management
- ✅ Ran `go mod tidy` to clean dependencies
- ✅ All builds successful (API + worker)
- ✅ All tests passing (60+ tests)
- ✅ Zero breaking changes
- ✅ Package analysis: All industry-standard packages optimal (chi, zap, viper, sqlx, jwt/v5)
- ✅ See `.claude/REFACTORING_ANALYSIS.md` and `.claude/CLEANUP_COMPLETE.md` for details
