# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Library Management System - A Go REST API service built with Clean Architecture principles. The project is optimized for vibecoding with fast feedback loops and clear separation of concerns.

**Module Name:** `library-service`
**Go Version:** 1.25.0
**Architecture:** Clean Architecture (Onion/Hexagonal)

## Essential Commands

### Quick Start
```bash
# Initialize and start development environment
make init                # Download dependencies
make up                  # Start PostgreSQL and Redis via Docker
make migrate-up          # Run store migrations
make run                 # Start API server (port 8080)

# Or use the combined dev command
make dev                 # Runs: up → migrate-up → run
```

### Development Workflow
```bash
# Core development commands
make test                # Run all tests (< 2 seconds)
make test-unit           # Unit tests only (with -short flag)
make test-integration    # Integration tests only
make test-coverage       # Generate HTML coverage report

# Single test execution
go test -v -run TestSpecificTest ./internal/domain/book

# Code quality
make fmt                 # Format code
make vet                 # Run go vet
make lint                # Run golangci-lint (25+ linters)
make ci                  # Full CI pipeline: fmt → vet → lint → test → build

# Building
make build               # Build all binaries (api, worker, migrate)
make build-api           # Build API server only
CGO_ENABLED=0 go build -o /tmp/library-api ./cmd/api     # Quick build to temp
CGO_ENABLED=0 go build -o /tmp/library-worker ./cmd/worker  # Quick worker build
CGO_ENABLED=0 go build -o /tmp/library-migrate ./cmd/migrate # Quick migrate build
./scripts/build.sh       # Build all with version info
```

### Database Operations
```bash
# Migrations
make migrate-create name=add_new_table  # Create new migration
make migrate-up                         # Apply pending migrations
make migrate-down                       # Rollback last migration
go run cmd/migrate/main.go up          # Direct migration command
```

### Docker Development
```bash
make up                  # Start services (PostgreSQL, Redis)
make down                # Stop services
make docker-logs         # View container logs
make restart             # Restart all services
make docker-build        # Build Docker images
```

### Other Useful Commands
```bash
# Tools and generators
make install-tools       # Install golangci-lint, mockgen, swag
make gen-mocks          # Generate mocks with go generate
make gen-docs           # Generate Swagger API documentation

# Module management
make mod-tidy           # Tidy and vendor go modules
make mod-update         # Update all dependencies

# Quality checks
make check              # Run fmt, vet, lint together
make security           # Run gosec security checks
make benchmark          # Run performance benchmarks
make version            # Show version information
```

## High-Level Architecture

The codebase follows Clean Architecture with strict dependency rules:

```
Domain → Use Case → Adapters → Infrastructure
(inner)                           (outer)
```

### Layer Structure

```
internal/
├── domain/              # Core business logic (zero dependencies)
│   ├── book/           # Book entity, service, repository interface
│   ├── member/         # Member entity, service, repository interface
│   └── author/         # Author entity, repository interface
│
├── usecase/            # Application business rules
│   ├── book/          # Book use cases (CreateBook, UpdateBook, etc.)
│   └── subscription/  # Subscription use cases
│
├── adapters/           # External interfaces
│   ├── http/          # HTTP handlers (REST API)
│   ├── repository/    # Database implementations
│   │   ├── postgres/  # PostgreSQL implementations
│   │   ├── mongo/     # MongoDB implementations
│   │   ├── memory/    # In-memory implementations
│   │   └── mocks/     # Mock implementations for testing
│   ├── cache/         # Redis cache implementations
│   ├── grpc/          # gRPC server
│   ├── email/         # SMTP email adapter
│   └── payment/       # Stripe/PayPal adapters
│
└── infrastructure/     # Technical concerns
    ├── auth/          # JWT authentication
    ├── config/        # Environment configuration
    ├── store/         # Database and cache store management
    ├── log/           # Zap structured logging
    ├── server/        # Server configuration
    └── app/           # Application initialization
```

### Key Architectural Patterns

1. **Repository Pattern**: Interfaces in domain layer, implementations in adapters
2. **Domain Services**: Complex business logic encapsulated in service objects
3. **Use Case Per Operation**: Single responsibility, one use case per business operation
4. **Constructor Dependency Injection**: All dependencies passed via constructors
5. **DTO Pattern**: Separate data transfer objects for each layer

## Domain-Driven Design Concepts

### Domain Entities
Each domain has specific typed entities (refactored from generic "Entity"):

```go
// internal/domain/book/entity.go
type Book struct {
    ID      string
    Name    *string
    Genre   *string
    ISBN    *string
    Authors []string
}

// internal/domain/member/entity.go
type Member struct {
    ID       string
    FullName *string
    Books    []string
}

// internal/domain/author/entity.go
type Author struct {
    ID        string
    FullName  *string
    Pseudonym *string
    Specialty *string
}
```

### Domain Services
Each domain has a service for business logic that doesn't belong to a single entity:

```go
// internal/domain/book/service.go
type Service struct{}

func (s *Service) ValidateISBN(isbn string) error     // ISBN validation logic
func (s *Service) ValidateBook(book Book) error       // Book validation
func (s *Service) CanBookBeDeleted(book Book) error   // Business rule check
```

### Repository Interfaces
Defined in domain, implemented in adapters:

```go
// internal/domain/book/repository.go
type Repository interface {
    List(ctx context.Context) ([]Book, error)
    Add(ctx context.Context, data Book) (string, error)
    Get(ctx context.Context, id string) (Book, error)
    Update(ctx context.Context, id string, data Book) error
    Delete(ctx context.Context, id string) error
}
```

### Use Cases
One file per operation, clear naming:

```go
// internal/usecase/book/create_book.go
type CreateBookUseCase struct {
    repo    book.Repository
    service *book.Service
}
```

## Testing Patterns

### Table-Driven Tests
```go
tests := []struct {
    name      string
    input     string
    wantError bool
}{
    {"valid case", "input", false},
    {"error case", "bad", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Test Organization
- `*_test.go` files alongside implementation
- `//go:build integration` tag for integration tests
- Mock generation: `go generate ./...`
- Test fixtures in `test/fixtures/`
- Benchmark files: `*_benchmark_test.go`
- Run benchmarks: `make benchmark`

## Key Dependencies

The project uses the following main dependencies:
- **HTTP Router**: `github.com/go-chi/chi/v5` - Lightweight, composable router
- **Database**: `github.com/jmoiron/sqlx` with `github.com/lib/pq` (PostgreSQL)
- **Migrations**: `github.com/golang-migrate/migrate/v4`
- **Validation**: `github.com/go-playground/validator/v10`
- **Logging**: `go.uber.org/zap` - Structured logging
- **Config**: `github.com/kelseyhightower/envconfig` - Environment-based config
- **Cache**: `github.com/redis/go-redis/v9` - Redis client
- **Testing**: `github.com/stretchr/testify` - Testing toolkit
- **MongoDB**: `go.mongodb.org/mongo-driver` - MongoDB driver (optional)
- **gRPC**: `google.golang.org/grpc` - gRPC server/client

## Important Configuration

### Environment Variables
```bash
# Copy and edit .env file
cp .env.example .env

# Key variables:
DATABASE_URL=postgres://user:pass@localhost/library
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
LOG_LEVEL=debug
```

### Database Connection
- PostgreSQL 15+ required
- Connection pooling: max 25 connections
- Migrations in `migrations/` directory

### API Documentation
- OpenAPI spec: `api/openapi/swagger.yaml`
- Swagger UI: http://localhost:8080/swagger/ (when running)
- Generate docs: `make gen-docs`

## Recent Refactoring (Google Go Style Guide Compliance)

The codebase was refactored to follow Google Go Style Guide best practices:

### Type System Changes
- **Before**: Generic `Entity` type in all domains
- **After**: Specific types (`Book`, `Member`, `Author`)
- All ~60 files updated with proper type references
- DTO functions renamed: `ParseFromEntity` → `ParseFromBook/Member/Author`

### File Organization Changes
- `database/` → `store/` (clearer naming)
- `logger/` → `log/` (Go convention)
- `mock/` → `mocks/` (Go convention)
- HTTP files simplified: `book_handler.go` → `book.go`
- DTOs simplified: `book_dto.go` → `book.go`

### Documentation Added
- Package-level `doc.go` files for all packages:
  - `internal/domain/doc.go` - Core domain layer overview
  - `internal/domain/book/doc.go` - Book domain with ISBN validation
  - `internal/domain/member/doc.go` - Member and subscription management
  - `internal/domain/author/doc.go` - Author management
  - `internal/usecase/doc.go` - Use case layer
  - `internal/adapters/doc.go` - Adapters layer
  - `internal/infrastructure/doc.go` - Infrastructure layer
  - `pkg/doc.go` - Shared utilities
- Comprehensive godoc comments with usage examples
- Architecture documentation in domain layer

## Code Standards

### File Organization
- One use case per file
- File size < 300 lines (max 500)
- Cyclomatic complexity < 10
- Package documentation in `doc.go` files

### Error Handling
```go
// Wrap errors with context
return fmt.Errorf("failed to create book: %w", err)

// Custom domain errors
return errors.ErrInvalidISBN
```

### Naming Conventions (Google Go Style Guide)
- **Domain entities**: Specific types (`Book`, `Member`, `Author`) not generic `Entity`
- **Use cases**: `CreateBookUseCase`, `UpdateMemberUseCase`
- **Services**: `Service` (package-scoped in each domain package)
- **Repositories**: `Repository` interface, implementations like `PostgresRepository`
- **Files**: Snake case for multi-word concepts (`create_book.go`, `subscribe_member.go`)
- **Imports**: Organized in groups (standard → external → internal) with blank lines
- **Package documentation**: All packages have `doc.go` files with godoc comments

## Performance Considerations

### Build Optimization
- Build time < 5 seconds
- Binary size ~15-20MB with CGO_ENABLED=0
- Use `CGO_ENABLED=0` for production builds

### Caching Strategy
- Redis for frequently accessed data
- 5-minute TTL for read cache
- Write-through cache updates

### Database Queries
- Use prepared statements
- Implement pagination for lists
- Add indexes on frequently queried columns

## Common Tasks

### Adding a New Domain
1. Create domain folder: `internal/domain/newdomain/`
2. Define entity, service, repository interface
3. Create use cases in `internal/usecase/newdomain/`
4. Implement repository in `internal/adapters/repository/`
5. Add HTTP handlers in `internal/adapters/http/`
6. Wire dependencies in `cmd/api/main.go`

### Running Background Jobs
```bash
make run-worker          # Start worker process
go run cmd/worker/main.go
```

### Debugging
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run with race detector
go test -race ./...

# Profile benchmarks
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# VSCode debugging - 5 launch configurations available:
# - Debug API Server
# - Debug Worker
# - Debug Migration
# - Debug Current Test
# - Debug Current File
```

## Project-Specific Patterns

### Dependency Injection Setup
All dependencies are wired manually in `cmd/api/main.go` using constructor injection. No DI framework is used.

### Request Flow
1. HTTP Request → Handler (adapters/http)
2. Handler → Use Case (usecase)
3. Use Case → Domain Service (domain)
4. Use Case → Repository (adapters/repository)
5. Response flows back in reverse

### Transaction Handling
Use cases define transaction boundaries. Repositories should not manage transactions.

### Validation Strategy
- Input validation in handlers
- Business rule validation in domain services
- Use `github.com/go-playground/validator/v10` for struct validation

### Linting Configuration
The project uses `.golangci.yml` with 25+ linters:
- **Core**: gofmt, goimports, govet, errcheck, staticcheck, unused
- **Quality**: misspell, unconvert, prealloc, nakedret, gocritic, revive
- **Security**: gosec
- **Complexity**: gocyclo (max: 10), gocognit (max: 20)
- **Duplication**: dupl (threshold: 100)
- **Error handling**: nilerr, wrapcheck
- **SQL**: sqlclosecheck, rowserrcheck
- **Context**: noctx, contextcheck

## CI/CD Pipeline

GitHub Actions workflows available:
- **ci.yml**: Main CI pipeline (lint, test, build, security scan)
- **claude-code-review.yml**: Automated code review
- **claude.yml**: PR assistant

Workflow runs on every push:
1. Lint with golangci-lint
2. Run tests with coverage
3. Build binaries
4. Security scan with gosec

Local CI simulation: `make ci`

## Troubleshooting

### Common Issues
- **Import cycles**: Check dependency direction (domain ← usecase ← adapters)
- **Test failures**: Ensure database is migrated (`make migrate-up`)
- **Lint errors**: Run `make fmt` before `make lint`
- **Connection refused**: Check if PostgreSQL/Redis are running (`make up`)
- **Module not found**: Run `go mod tidy` or `make mod-tidy`
- **Build failures**: Ensure Go 1.25 is installed

### Debug Commands
```bash
# Check running services
docker-compose -f deployments/docker/docker-compose.yml ps

# Database connection test
psql $DATABASE_URL -c "SELECT 1"

# Redis connection test
redis-cli ping
```

## Additional Documentation

The project includes extensive documentation for deeper understanding:

### Getting Started Guides
- **[Quick Start (5 min)](./docs/guides/QUICKSTART.md)** - Get running in 5 minutes
- **[Development Guide](./docs/guides/DEVELOPMENT.md)** - Comprehensive workflow
- **[Contributing Guidelines](./docs/guides/CONTRIBUTING.md)** - How to contribute

### Architecture Documentation
- **[Architecture Overview](./docs/architecture.md)** - System design principles
- **[Package Overview](./docs/package-overview.md)** - Package structure and dependencies
- **[Architecture Decisions (ADRs)](./docs/adr/README.md)** - Key architectural decisions
  - [ADR-001: Clean Architecture](./docs/adr/001-clean-architecture.md)
  - [ADR-002: Domain Services](./docs/adr/002-domain-services.md)
  - [ADR-003: Dependency Injection](./docs/adr/003-dependency-injection.md)

### Code Examples
- **[Basic CRUD](./examples/basic_crud/)** - Complete CRUD workflow example
- **[Domain Services](./examples/domain_service/)** - Business logic patterns
- **[Testing Patterns](./examples/testing/)** - Testing strategies and mocks

### Layer-Specific READMEs
- **[Domain Layer](./internal/domain/README.md)** - Business logic and entities
- **[Use Case Layer](./internal/usecase/README.md)** - Application use cases
- **[Adapter Layer](./internal/adapters/README.md)** - External interfaces
- **[Command Line Apps](./cmd/README.md)** - Entry points (API, worker, migrate)
- **[Shared Packages](./pkg/README.md)** - Reusable utilities

### Testing Resources
- **[Test Fixtures](./test/fixtures/README.md)** - Shared test data and helpers
- **Integration Tests**: `./test/integration/` with `//go:build integration` tag
- **Benchmarks**: `*_benchmark_test.go` files throughout the codebase