# Development Guide

> **Complete setup, commands, and daily development workflow**

## Quick Start

### Option A: Automated Setup (Recommended - 5-10 Minutes)

```bash
# Run automated setup script (does everything for you)
./scripts/dev-setup.sh
```

This single command will:
- ✅ Check prerequisites (Go, Docker, Make)
- ✅ Install dependencies and development tools
- ✅ Configure git hooks for quality checks
- ✅ Start Docker services (PostgreSQL, Redis)
- ✅ Run database migrations
- ✅ Seed development data (test users & books)
- ✅ Build the project
- ✅ Verify everything works

**Test Accounts** (created by seed script):
- `admin@library.com` / `Admin123!@#`
- `user@library.com` / `User123!@#`
- `premium@library.com` / `Premium123!@#`

### Option B: Manual Setup (5 Minutes)

```bash
# 1. Install dependencies
make init

# 2. Start Docker service (PostgreSQL + Redis)
make up

# 3. Run database migrations
make migrate-up

# 4. Install git hooks (optional but recommended)
make install-hooks

# 5. Seed development data (optional)
./scripts/seed-data.sh

# 6. Start API server
make run
```

**Verification:**
- API running at http://localhost:8080
- Swagger docs at http://localhost:8080/swagger/index.html
- Database at postgres://library:library123@localhost:5432/library

## Prerequisites

### Required
- **Go 1.25+** - [Download](https://go.dev/dl/)
- **Docker & Docker Compose** - [Download](https://www.docker.com/get-started)
- **Make** - Usually pre-installed on macOS/Linux
- **Git** - For version control

### Optional (for advanced development)
- **golangci-lint** - Linting (installed via `make install-tools`)
- **swag** - Swagger generation (installed via `make install-tools`)
- **psql** - PostgreSQL CLI client
- **redis-cli** - Redis CLI client

## Project Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd library
```

### 2. Install Development Tools

```bash
# Install golangci-lint, swag, and other tools
make install-tools
```

This installs:
- `golangci-lint` → Code linting (25+ linters)
- `swag` → Swagger/OpenAPI documentation generation
- `staticcheck` → Advanced static analysis
- `goimports` → Import formatting

### 3. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

**Critical Environment Variables:**

```bash
# Application
APP_MODE=dev                    # dev (verbose logs) or prod (JSON logs)
APP_PORT=8080                   # API server port

# Database
POSTGRES_DSN="postgres://library:library123@localhost:5432/library?sslmode=disable"
# Format: postgres://user:password@host:port/database?sslmode=disable

# Cache (optional - falls back to memory cache if not available)
REDIS_HOST=localhost:6379
REDIS_PASSWORD=                 # Empty for local development
REDIS_DB=0

# Authentication
JWT_SECRET="your-secret-key-change-in-production"  # **REQUIRED** - Change in production!
JWT_EXPIRY=24h                  # Access token expiry (24 hours)

# Payment Gateway (epayment.kz)
EPAYMENT_BASE_URL="https://api.epayment.kz"
EPAYMENT_CLIENT_ID="your-client-id"
EPAYMENT_CLIENT_SECRET="your-client-secret"
EPAYMENT_TERMINAL="your-terminal-id"
EPAYMENT_WIDGET_URL="https://widget.epayment.kz"

# Worker (background jobs)
PAYMENT_EXPIRY_INTERVAL=5m      # How often to check for expired payments
CALLBACK_RETRY_INTERVAL=1m      # How often to retry failed callbacks
```

### 4. Start Dependencies

```bash
# Start PostgreSQL and Redis in Docker
make up

# Verify service are running
docker-compose -f deployments/docker/docker-compose.yml ps
```

Expected output:
```
library-postgres   Up   0.0.0.0:5432->5432/tcp
library-redis      Up   0.0.0.0:6379->6379/tcp
```

### 5. Run Database Migrations

```bash
# Apply all migrations
make migrate-up

# Check current version
psql "postgres://library:library123@localhost:5432/library" -c "SELECT version FROM schema_migrations"
```

### 6. Verify Setup

```bash
# Run tests to verify everything works
make test

# Build binaries
make build

# Start API server
make run
```

Navigate to http://localhost:8080/swagger/index.html - you should see Swagger UI.

## Common Commands

### Development Workflow

```bash
# Start full development environment (recommended)
make dev                  # Starts docker + migrations + API server

# Or run individually:
make up                   # Start docker service only
make migrate-up           # Run migrations
make run                  # Start API server
make run-worker           # Start background worker
```

### Building

```bash
make build                # Build all binaries (api, worker, migrate)
make build-api            # Build API server → bin/library-api
make build-worker         # Build worker → bin/library-worker
make build-migrate        # Build migration tool → bin/library-migrate
```

Binaries are output to `bin/` directory:
```
bin/
├── library-api           # HTTP API server
├── library-worker        # Background worker
└── library-migrate       # Migration runner
```

### Testing

```bash
make test                 # All tests with race detection + coverage
make test-unit            # Unit tests only (fast, no database)
make test-integration     # Integration tests (requires database)
make test-coverage        # Generate HTML coverage report → coverage.html
```

**Run specific tests:**
```bash
# Run all tests in a package
go test -v ./internal/domain/book/

# Run specific test function
go test -v -run TestBookService_ValidateISBN ./internal/domain/book/

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./internal/domain/book/
```

**Integration tests:**
```bash
# Requires TEST_POSTGRES_DSN environment variable
TEST_POSTGRES_DSN="postgres://library:library123@localhost:5432/library_test?sslmode=disable" make test-integration

# Or use default database
make test-integration
```

### Code Quality

```bash
make ci                   # Full CI pipeline: fmt → vet → lint → test → build
make fmt                  # Format code (gofmt + goimports)
make vet                  # Run go vet (suspicious constructs)
make lint                 # Run golangci-lint (25+ linters)
```

**Before committing, always run:**
```bash
make ci
```

This ensures:
- Code is formatted
- No vet issues
- All linters pass
- All tests pass
- Builds successfully

### Database Migrations

```bash
# Create new migration
make migrate-create name=add_book_ratings
# Creates:
# migrations/postgres/NNNNNN_add_book_ratings.up.sql
# migrations/postgres/NNNNNN_add_book_ratings.down.sql

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

**Direct usage:**
```bash
# Apply all pending migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down

# Create new migration
go run cmd/migrate/main.go create add_book_ratings

# With custom database
POSTGRES_DSN="postgres://user:pass@localhost:5432/db" go run cmd/migrate/main.go up
```

### API Documentation

```bash
# Regenerate Swagger documentation
make gen-docs

# Manual regeneration
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal
```

**Important**: Always regenerate docs after:
- Adding new endpoints
- Changing request/response DTOs
- Updating Swagger annotations

**View documentation:**
- Start server: `make run`
- Open browser: http://localhost:8080/swagger/index.html

### Docker Management

```bash
make up                   # Start service (PostgreSQL + Redis)
make down                 # Stop service
make logs                 # View service logs
make clean-docker         # Remove containers and volumes (destructive!)
```

**Individual service logs:**
```bash
docker-compose -f deployments/docker/docker-compose.yml logs -f postgres
docker-compose -f deployments/docker/docker-compose.yml logs -f redis
```

### Development Tools

```bash
make install-tools        # Install all development tools
make gen-mocks            # Generate test mocks (if using mockery)
make benchmark            # Run performance benchmarks
```

## Daily Development Workflow

### Starting Your Day

```bash
# 1. Pull latest changes
git pull origin main

# 2. Update dependencies
go mod download

# 3. Start service
make up

# 4. Apply any new migrations
make migrate-up

# 5. Run tests to ensure everything works
make test

# 6. Start development server
make dev
```

### During Development

```bash
# Auto-restart on code changes (use air or similar)
# For now, manually restart:
Ctrl+C
make run

# Run tests frequently
go test -v ./internal/domain/book/

# Check code quality
make lint
```

### Before Committing

```bash
# 1. Format code
make fmt

# 2. Run full CI pipeline
make ci

# 3. Regenerate docs if needed
make gen-docs

# 4. Commit
git add .
git commit -m "feat: add book rating feature"
git push
```

## Running the Application

### API Server

```bash
# Development mode (verbose logging)
APP_MODE=dev make run

# Production mode (JSON logging)
APP_MODE=prod make run

# With custom port
APP_PORT=9000 make run

# Direct execution
go run cmd/api/main.go
```

### Background Worker

```bash
# Start worker
make run-worker

# Direct execution
go run cmd/worker/main.go
```

The worker processes:
- **Payment Expiry Job** - Runs every 5 minutes, marks expired pending payments as failed
- **Callback Retry Job** - Runs every 1 minute, retries failed webhook callbacks

### Both Services

```bash
# Terminal 1: API server
make run

# Terminal 2: Background worker
make run-worker
```

## Testing the API

### Using cURL

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#",
    "full_name": "Test User"
  }'

# Login (get JWT token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!@#"
  }'

# Use token for authenticated requests
TOKEN="<access_token_from_login_response>"

curl -X GET http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN"
```

### Using Swagger UI

1. Start server: `make run`
2. Open browser: http://localhost:8080/swagger/index.html
3. Click "Authorize" button
4. Enter token: `Bearer <your_access_token>`
5. Try API endpoints interactively

### Using Postman

Import the OpenAPI spec: `api/openapi/swagger.json`

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -ti:8080

# Kill the process
lsof -ti:8080 | xargs kill -9

# Or use different port
APP_PORT=9000 make run
```

### Database Connection Errors

```bash
# Check if PostgreSQL is running
docker-compose -f deployments/docker/docker-compose.yml ps postgres

# View PostgreSQL logs
docker-compose -f deployments/docker/docker-compose.yml logs postgres

# Restart PostgreSQL
make down && make up

# Connect manually to verify
psql "postgres://library:library123@localhost:5432/library"
```

### Migration Errors

```bash
# Check current migration version
psql "postgres://library:library123@localhost:5432/library" \
  -c "SELECT * FROM schema_migrations"

# Force reset (DESTRUCTIVE - only in development!)
make migrate-down
make migrate-down
# ... repeat until all migrations rolled back
make migrate-up
```

### Test Failures

```bash
# Clear test cache
go clean -testcache

# Run with verbose output
go test -v ./...

# Run specific failing test
go test -v -run TestThatFails ./internal/domain/book/
```

### Build Errors

```bash
# Clean build cache
go clean -cache

# Update dependencies
go mod download
go mod tidy

# Rebuild
make build
```

## Performance Tips

### Speed Up Tests

```bash
# Run only unit tests (skip integration)
make test-unit

# Run tests in parallel (default)
go test -parallel 8 ./...

# Skip slow tests
go test -short ./...
```

### Speed Up Builds

```bash
# Use build cache (default)
go build -o bin/library-api ./cmd/api

# Disable CGO for faster builds (project default)
CGO_ENABLED=0 go build -o bin/library-api ./cmd/api
```

## Hot Reload / Live Reload

### Option 1: Using Air

```bash
# Install
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air

# Or create .air.toml config
air init
```

### Option 2: Using Reflex

```bash
# Install
go install github.com/cespare/reflex@latest

# Run (automatically rebuilds on changes)
reflex -r '\.go$' -s -- sh -c 'make build && ./bin/library-api'
```

### Option 3: Using Make Watch

```bash
# Add to Makefile:
watch:
    reflex -r '\.go$' -s -- sh -c 'make test && make run'

# Then use:
make watch
```

## Debugging

### VS Code Debugging

Create `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch API",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/api",
      "env": {
        "POSTGRES_DSN": "postgres://library:library123@localhost:5432/library?sslmode=disable",
        "JWT_SECRET": "development-secret-key"
      }
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/domain/book"
    }
  ]
}
```

### Delve (Command Line Debugger)

```bash
# Install
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API server
dlv debug ./cmd/api

# Debug tests
dlv test ./internal/usecase/bookops

# Attach to running process
dlv attach <PID>
```

### Logging for Debugging

```go
import "library-service/internal/infrastructure/log"

// In code
log.Info("Processing book", "book_id", bookID, "title", book.Title)
log.Error("Failed to create book", "error", err)
log.Debug("Cache hit", "key", cacheKey)
```

**Filter logs:**
```bash
# Show only errors
tail -f service.log | jq 'select(.level=="error")'

# Show specific field
tail -f service.log | jq 'select(.book_id=="123")'

# Follow logs in real-time
make run 2>&1 | grep "ERROR"
```

### Enable Debug Logging

```bash
# Start with debug mode
APP_MODE=dev LOG_LEVEL=debug make run
```

## Performance Optimization

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/domain/book/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/domain/book/
go tool pprof mem.prof

# Run server with profiling
go run ./cmd/api &
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Benchmarking

```go
// internal/domain/book/service_benchmark_test.go
func BenchmarkValidateISBN(b *testing.B) {
    svc := NewService()
    isbn := "978-0-306-40615-7"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = svc.ValidateISBN(isbn)
    }
}
```

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/domain/book/

# Compare before/after
go test -bench=. -benchmem ./... > old.txt
# Make changes
go test -bench=. -benchmem ./... > new.txt
benchcmp old.txt new.txt
```

## Database Workflows

### Viewing Database State

```bash
# Connect to database
psql -h localhost -U library -d library

# Useful queries
\dt                 # List tables
\d books            # Describe books table
SELECT * FROM books LIMIT 10;
SELECT COUNT(*) FROM members WHERE subscription_type = 'premium';
```

### Resetting Database

```bash
# Full reset
make down
docker volume rm library-postgres-data
make up
make migrate-up

# Or just migrations
make migrate-down
make migrate-up
```

### Seeding Test Data

Create `scripts/seed.sh`:
```bash
#!/bin/bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}'

# Get token and create books...
```

## Code Quality Automation

### Pre-commit Hooks

Create `.git/hooks/pre-commit`:
```bash
#!/bin/bash
make fmt
make lint
make test

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi
```

```bash
chmod +x .git/hooks/pre-commit
```

### Git Aliases

Add to `~/.gitconfig`:
```ini
[alias]
    st = status
    co = checkout
    br = branch
    ci = commit
    unstage = reset HEAD --
    last = log -1 HEAD
    visual = log --oneline --graph --decorate
```

## Productivity Tips

### Shell Aliases

Add to `~/.bashrc` or `~/.zshrc`:
```bash
# Library project shortcuts
alias ld='cd ~/projects/library'
alias ldev='cd ~/projects/library && make dev'
alias ltest='make test'
alias lci='make ci'
alias lauth='export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '\''{"email":"test@example.com","password":"Test123!@#"}'\'' | jq -r '\''.tokens.access_token'\'')'
```

### API Testing Scripts

Create `scripts/test-api.sh`:
```bash
#!/bin/bash
set -e

# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# Create book
BOOK=$(curl -s -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Book","isbn":"9780132350884","genre":"Tech"}')

echo "Created: $BOOK"

# List books
curl -s http://localhost:8080/api/v1/books | jq
```

### Editor Snippets

**VS Code** - Create `.vscode/go.code-snippets`:
```json
{
  "Use Case": {
    "prefix": "usecase",
    "body": [
      "type ${1:Operation}UseCase struct {",
      "\trepo ${2:entity}.Repository",
      "}",
      "",
      "func New${1}UseCase(repo ${2}.Repository) *${1}UseCase {",
      "\treturn &${1}UseCase{repo: repo}",
      "}",
      "",
      "func (uc *${1}UseCase) Execute(ctx context.Context, req ${1}Request) (*${2}.Entity, error) {",
      "\t$0",
      "\treturn nil, nil",
      "}"
    ]
  }
}
```

## Next Steps

You now know how to:
- ✅ Set up the project locally
- ✅ Run tests and checks
- ✅ Build and run the application
- ✅ Work with migrations
- ✅ Debug common issues

Next, read:
- `.claude/common-tasks.md` - Step-by-step guides for adding features
- `.claude/coding-standards.md` - Code style and conventions
- `.claude/README.md` - What to read for specific tasks (task-specific quick starts)
