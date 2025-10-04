# Development Guide

Comprehensive guide for developing the Library Management System with vibecoding workflow.

## Development Workflow

### Fast Feedback Loop

The project is optimized for instant feedback:

```bash
# Watch mode (recommended)
make dev              # Start services + API with hot reload

# Test-driven development
make test             # < 2 seconds execution
make test-coverage    # With HTML coverage report

# Code quality checks
make lint             # golangci-lint (25+ linters)
make vet              # Go vet
make fmt              # Format code
```

### Typical Development Cycle

1. **Write test** → 2. **Implement** → 3. **Run tests** → 4. **Lint** → 5. **Commit**

```bash
# Full CI pipeline locally
make ci               # Runs: fmt → vet → lint → test → build
```

## Project Architecture

### Clean Architecture Layers

```
Domain Layer (Core Business Logic)
    ↓ depends on
Use Case Layer (Application Logic)
    ↓ depends on
Adapters Layer (External Interfaces)
    ↓ depends on
Infrastructure Layer (Framework & Tools)
```

**Dependency Rule**: Inner layers never depend on outer layers.

### File Organization

```
internal/
├── domain/              # Business entities, rules, interfaces
│   ├── book/           # Book domain
│   │   ├── entity.go   # Book entity
│   │   ├── service.go  # Business rules (ISBN validation, etc.)
│   │   └── repository.go  # Repository interface
│   └── member/         # Member domain (similar structure)
│
├── usecase/            # Application use cases
│   ├── book/
│   │   ├── create_book.go      # One file per use case
│   │   ├── create_book_test.go # Test file
│   │   └── dto.go              # Use case DTOs
│   └── member/
│
└── adapters/           # External interfaces
    ├── http/           # HTTP handlers
    ├── repository/     # Database implementations
    └── cache/          # Cache implementations
```

## Adding New Features

### 1. Domain-First Approach

Start with the domain layer:

```go
// 1. Define entity (internal/domain/newdomain/entity.go)
type NewEntity struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

// 2. Create domain service (internal/domain/newdomain/service.go)
type Service struct{}

func NewService() *Service {
    return &Service{}
}

func (s *Service) ValidateEntity(entity Entity) error {
    // Business rules here
    if entity.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

// 3. Define repository interface (internal/domain/newdomain/repository.go)
type Repository interface {
    Create(ctx context.Context, entity Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
}
```

### 2. Use Case Layer

```go
// internal/usecase/newdomain/create_entity.go
type CreateEntityUseCase struct {
    repo    domain.Repository
    service *domain.Service
}

func (uc *CreateEntityUseCase) Execute(ctx context.Context, input CreateEntityInput) error {
    // 1. Map DTO to entity
    entity := domain.Entity{
        ID:   uuid.New().String(),
        Name: input.Name,
    }

    // 2. Validate with domain service
    if err := uc.service.ValidateEntity(entity); err != nil {
        return err
    }

    // 3. Persist
    return uc.repo.Create(ctx, entity)
}
```

### 3. Adapter Layer

```go
// internal/adapters/http/newdomain/handler.go
type Handler struct {
    createUC *usecase.CreateEntityUseCase
}

func (h *Handler) Create(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := h.createUC.Execute(c.Request.Context(), req.ToInput()); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"success": true})
}
```

### 4. Wire It Up

```go
// cmd/api/main.go or dependency injection setup
domainService := newdomain.NewService()
repo := newdomainrepo.NewPostgresRepository(db)
createUC := usecase.NewCreateEntityUseCase(repo, domainService)
handler := newdomainhttp.NewHandler(createUC)

// Register routes
router.POST("/newdomain", handler.Create)
```

## Testing Strategy

### Test Pyramid

```
        E2E (Few)
       /         \
    Integration (Some)
   /                  \
  Unit Tests (Many - 80%+)
```

### Unit Tests (Domain & Use Cases)

```go
// internal/domain/book/service_test.go
func TestService_ValidateISBN(t *testing.T) {
    service := NewService()

    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-10", "0-306-40615-2", false},
        {"invalid ISBN", "invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateISBN() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```bash
# Run with database
DB_HOST=localhost make test-integration
```

```go
// test/integration/book_test.go
//go:build integration

func TestBookRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := repository.NewBookRepository(db)

    book := domain.Book{ID: "test-1", Title: "Test"}
    err := repo.Create(context.Background(), book)

    assert.NoError(t, err)
}
```

### Benchmarks

```bash
make benchmark        # Run all benchmarks
```

```go
// internal/domain/book/service_benchmark_test.go
func BenchmarkService_ValidateISBN(b *testing.B) {
    service := NewService()
    isbn := "978-0-306-40615-7"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = service.ValidateISBN(isbn)
    }
}
```

## Code Quality

### Linting Configuration

`.golangci.yml` enables 25+ linters:

- **Security**: gosec
- **Complexity**: gocyclo (max 10), gocognit (max 20)
- **Errors**: errcheck, wrapcheck, nilerr
- **Style**: gofmt, goimports, revive
- **Duplication**: dupl (threshold 100)

```bash
make lint             # Run all linters
make lint-fix         # Auto-fix issues (if supported)
```

### Code Standards

**File Size**: Keep files < 300 lines (max 500)
**Function Complexity**: Cyclomatic < 10, Cognitive < 20
**Test Coverage**: Domain services 100%, Use cases 80%+, Overall 60%+

## Debugging

### VSCode Debug Configuration

`.vscode/launch.json`:

```json
{
  "configurations": [
    {
      "name": "Debug API",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/api"
    }
  ]
}
```

### Debugging Tips

```bash
# Enable debug logging
export LOG_LEVEL=debug
make run

# Run specific test with verbose output
go test -v -run TestSpecificTest ./internal/domain/book

# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

## Database Migrations

### Create Migration

```bash
make migrate-create name=add_new_table
# Creates: migrations/000X_add_new_table.up.sql
#          migrations/000X_add_new_table.down.sql
```

### Apply Migrations

```bash
make migrate-up       # Apply all pending
make migrate-down     # Rollback last migration
```

## Docker Development

### Local Development

```bash
make up               # Start services (PostgreSQL, Redis)
make down             # Stop services
make docker-logs      # View logs
```

### Full Docker Build

```bash
make docker-build     # Build all images (api, worker, migrate)
```

## Performance Optimization

### Build Optimization

```bash
# Production build (stripped, optimized)
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api

# Or use Makefile
make build-api
```

### Database Optimization

- Use indexes on frequently queried columns
- Implement pagination for list endpoints
- Use Redis caching for read-heavy operations
- Batch operations where possible

## Best Practices for Vibecoding

1. **Single Responsibility**: One use case per file, one method per business operation
2. **Testability**: Write tests before implementation (TDD)
3. **Clear Naming**: `CreateBookUseCase`, `BookService`, `BookRepository`
4. **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
5. **Interface Segregation**: Keep interfaces small and focused
6. **Dependency Injection**: Constructor injection pattern
7. **Immutability**: Prefer value objects, avoid shared mutable state
8. **Documentation**: Package-level comments, godoc-friendly

## Troubleshooting

### Common Issues

**Import cycle**: Check dependency direction (domain ← usecase ← adapters)
**Test failures**: Ensure test database is migrated (`make migrate-up`)
**Lint errors**: Run `make fmt` before `make lint`
**Build slow**: Use `go build -i` to cache intermediate results

### Getting Help

1. Check [Architecture docs](../architecture.md)
2. Review [ADRs](../adr/) for design decisions
3. Look at existing domain for patterns
4. Check test files for usage examples

## CI/CD Pipeline

GitHub Actions runs on every push:

1. **Lint** → golangci-lint
2. **Test** → Unit + integration with coverage
3. **Build** → Multi-platform binaries
4. **Security** → gosec + govulncheck
5. **Quality** → SonarCloud (if configured)

Local CI simulation:

```bash
make ci               # Run full pipeline locally
```

## Next Steps

- Read [Contributing Guidelines](./CONTRIBUTING.md)
- Explore [Example Code](../../examples/)
- Review [Architecture Decisions](../adr/)
