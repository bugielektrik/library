# Essential Commands Reference

> **All commands you need for development, testing, and deployment**

## Quick Start

```bash
# Initialize and start development environment
make init                # Download dependencies
make up                  # Start PostgreSQL and Redis via Docker
make migrate-up          # Run database migrations
make run                 # Start API server (port 8080)

# Or use the combined dev command
make dev                 # Runs: up → migrate-up → run
```

## Development Workflow

### Testing
```bash
make test                # Run all tests (< 2 seconds)
make test-unit           # Unit tests only (with -short flag)
make test-integration    # Integration tests only
make test-coverage       # Generate HTML coverage report

# Single test execution
go test -v -run TestSpecificTest ./internal/domain/book
go test -v ./internal/usecase/book/...

# Watch mode (install first: go install github.com/cespare/reflex@latest)
reflex -r '\.go$' -s -- sh -c 'go test ./...'
```

### Code Quality
```bash
make fmt                 # Format code (gofmt + goimports)
make vet                 # Run go vet
make lint                # Run golangci-lint (25+ linters)
make ci                  # Full CI pipeline: fmt → vet → lint → test → build
make check               # Quick check: fmt → vet → lint
make security            # Run gosec security checks
```

### Building
```bash
make build               # Build all binaries (api, worker, migrate)
make build-api           # Build API server only → bin/library-api
make build-worker        # Build worker only → bin/library-worker
make build-migrate       # Build migration tool → bin/library-migrate

# Quick temporary builds (faster, no version info)
CGO_ENABLED=0 go build -o /tmp/library-api ./cmd/api
CGO_ENABLED=0 go build -o /tmp/library-worker ./cmd/worker
CGO_ENABLED=0 go build -o /tmp/library-migrate ./cmd/migrate
```

## Database Operations

### Migrations
```bash
make migrate-create name=add_new_table  # Create new migration
make migrate-up                         # Apply pending migrations
make migrate-down                       # Rollback last migration

# Direct migration commands (requires POSTGRES_DSN env var)
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
go run cmd/migrate/main.go create add_new_column

# With custom DSN
POSTGRES_DSN="postgres://user:pass@localhost:5432/library?sslmode=disable" \
  go run cmd/migrate/main.go up
```

### Database Access
```bash
# Via Docker
docker exec -it library-postgres psql -U library -d library

# Direct connection
psql -h localhost -U library -d library

# Useful queries
psql -U library -d library -c "SELECT * FROM books LIMIT 5;"
psql -U library -d library -c "\dt"  # List tables
```

## Docker Operations

```bash
# Core commands
make up                  # Start services (PostgreSQL, Redis)
make down                # Stop services
make docker-logs         # View container logs
make restart             # Restart all services
make docker-build        # Build Docker images

# Direct docker-compose commands
cd deployments/docker
docker-compose up -d
docker-compose down
docker-compose logs -f api
docker-compose ps

# Clean everything
docker-compose down -v   # Remove volumes (resets database)
```

## Development Tools

```bash
# Install development tools
make install-tools       # Install golangci-lint, mockgen, swag

# Generate code
make gen-mocks          # Generate mocks with go generate
make gen-docs           # Generate Swagger API documentation

# Module management
make mod-tidy           # Tidy and vendor go modules
make mod-update         # Update all dependencies
go mod download         # Download dependencies
go mod vendor           # Vendor dependencies
```

## Testing & Debugging

```bash
# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detection
go test -race ./...

# Benchmarks
make benchmark
go test -bench=. -benchmem ./internal/domain/book/

# Verbose output
go test -v ./internal/usecase/book/

# Run specific test
go test -v -run TestCreateBook ./internal/usecase/book/
```

## API Testing

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}'

# Login and get token
export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' | jq -r '.tokens.access_token')

# Create book (authenticated)
curl -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Clean Code","isbn":"9780132350884","genre":"Tech","authors":[]}'

# List books
curl http://localhost:8080/api/v1/books | jq
```

## Utility Commands

```bash
# Clean build artifacts
make clean

# Show version info
make version

# Check for security issues
gosec ./...

# Find TODO comments
grep -r "TODO" --include="*.go" ./internal

# Find large files
find . -type f -name "*.go" -exec wc -l {} + | sort -rn | head -10

# Dependency graph
go mod graph | grep library-service

# Check unused dependencies
go mod tidy -v
```

## Performance & Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/domain/book/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/domain/book/
go tool pprof mem.prof

# Build with profiling enabled
go build -o api-profiled ./cmd/api
./api-profiled  # Then visit http://localhost:6060/debug/pprof/
```

## Deployment

```bash
# Build for production
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/library-api ./cmd/api

# Build all platforms
GOOS=linux GOARCH=amd64 make build    # Linux
GOOS=darwin GOARCH=amd64 make build   # macOS Intel
GOOS=darwin GOARCH=arm64 make build   # macOS Apple Silicon
GOOS=windows GOARCH=amd64 make build  # Windows

# Docker build
docker build -t library-api:latest -f deployments/docker/Dockerfile .
docker run -p 8080:8080 library-api:latest
```

## Emergency Commands

```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Reset everything
make down && make up && make migrate-up

# Clear test cache
go clean -testcache

# Fix import issues
goimports -w .
go mod tidy

# Rebuild from scratch
make clean && make build

# Reset database
docker-compose down -v
docker-compose up -d
make migrate-up
```

## All Make Commands

```bash
make help                # Show all available commands
make init                # Initialize project dependencies
make build               # Build all binaries
make build-api           # Build API server binary
make build-worker        # Build worker binary
make build-migrate       # Build migration tool binary
make run                 # Run API server locally
make run-worker          # Run worker locally
make test                # Run all tests
make test-unit           # Run unit tests only
make test-integration    # Run integration tests only
make test-coverage       # Run tests with coverage report
make lint                # Run linters
make fmt                 # Format code
make vet                 # Run go vet
make clean               # Clean build artifacts
make migrate-up          # Run database migrations
make migrate-down        # Rollback database migrations
make migrate-create      # Create new migration (usage: make migrate-create name=migration_name)
make up                  # Start Docker services
make down                # Stop Docker services
make docker-logs         # Show logs from docker-compose services
make docker-build        # Build Docker images
make restart             # Restart docker services
make install-tools       # Install development tools
make gen-mocks           # Generate mocks for testing
make gen-docs            # Generate API documentation
make dev                 # Start development environment
make ci                  # Run CI pipeline locally
make check               # Run all checks (format, vet, lint)
make mod-tidy            # Tidy go modules
make mod-update          # Update go modules
make benchmark           # Run benchmarks
make security            # Run security checks
make version             # Show version information
```
