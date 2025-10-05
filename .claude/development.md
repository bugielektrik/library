# Development Workflow Guide

> **Daily development tasks, tips, and best practices**

## Daily Workflow

### Morning Routine

```bash
# 1. Pull latest changes
git pull origin main

# 2. Update dependencies if needed
go mod tidy

# 3. Start infrastructure
make up

# 4. Run migrations (if new)
make migrate-up

# 5. Start development server
make dev

# 6. Verify everything works
make test
```

### Feature Development Workflow

**1. Create Feature Branch**
```bash
git checkout -b feature/add-book-ratings
```

**2. Write Tests First (TDD)**
```bash
# Create test file
touch internal/domain/rating/service_test.go

# Write failing tests
go test ./internal/domain/rating/
```

**3. Implement Feature**

Follow the layer order:
1. **Domain** (`internal/domain/rating/`)
   - `entity.go` - Define Rating entity
   - `service.go` - Business rules
   - `repository.go` - Interface
   - `service_test.go` - Unit tests (100% coverage)

2. **Use Case** (`internal/usecase/rating/`)
   - `add_rating.go` - Add rating use case
   - `get_ratings.go` - Get ratings use case
   - Tests with mocked repositories

3. **Adapter** (`internal/adapters/`)
   - `repository/postgres/rating.go` - PostgreSQL implementation
   - `http/handlers/rating.go` - HTTP handlers
   - `http/dto/rating.go` - Request/response DTOs

4. **Wire Dependencies** (`internal/usecase/container.go`)
   ```go
   type Container struct {
       // ... existing use cases
       AddRating  *rating.AddRatingUseCase
       GetRatings *rating.GetRatingsUseCase
   }
   ```

5. **Database Migration**
   ```bash
   make migrate-create name=create_ratings_table
   # Edit migration files
   make migrate-up
   ```

**4. Test Everything**
```bash
make test           # Unit tests
make test-coverage  # Check coverage
make lint           # Code quality
```

**5. Manual Testing**
```bash
# Start server
make run

# Test API endpoint
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"book_id":"123","rating":5,"comment":"Great book!"}'
```

**6. Commit & Push**
```bash
git add .
git commit -m "feat: add book rating system"
git push origin feature/add-book-ratings
```

## Hot Reload / Live Reload

**Option 1: Using Air**
```bash
# Install
go install github.com/cosmtrek/air@latest

# Run
air

# Or create .air.toml config
air init
```

**Option 2: Using Reflex**
```bash
# Install
go install github.com/cespare/reflex@latest

# Run (automatically rebuilds on changes)
reflex -r '\.go$' -s -- sh -c 'make build && ./bin/library-api'
```

**Option 3: Using Make Watch**
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
dlv test ./internal/usecase/book

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
alias ltest='cd ~/projects/library && make test'
alias llog='tail -f ~/projects/library/service.log'

# Quick auth
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

## Troubleshooting

### Common Issues

**"Port already in use"**
```bash
lsof -ti:8080 | xargs kill -9
```

**"Database connection refused"**
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Restart
make down && make up
sleep 5  # Wait for startup
```

**"Import cycle not allowed"**
- Domain importing from adapters → Fix dependency direction
- Check with: `go list -f '{{ .ImportPath }}' ./...`

**"Test fails intermittently"**
```bash
# Race condition likely
go test -race ./...

# Run multiple times
go test -count=100 ./internal/usecase/book/
```

**"Linter errors after refactoring"**
```bash
go mod tidy
goimports -w .
make fmt
```

## Next Steps

- Review [Testing Guide](./testing.md) for testing strategies
- Check [API Documentation](./api.md) for endpoint details
- See [Standards](./standards.md) for code conventions
