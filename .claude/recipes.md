# Quick Recipes

> **Common tasks solved with copy-paste code snippets**

## Database Operations

### Create Migration

```bash
# Create a new migration
make migrate-create name=add_loans_table

# This creates:
# migrations/postgres/XXXXXX_add_loans_table.up.sql
# migrations/postgres/XXXXXX_add_loans_table.down.sql
```

### Sample Migration

```sql
-- migrations/postgres/000004_create_loans_table.up.sql
CREATE TABLE IF NOT EXISTS loans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    loan_date TIMESTAMP NOT NULL DEFAULT NOW(),
    due_date TIMESTAMP NOT NULL,
    return_date TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (status IN ('active', 'returned', 'overdue'))
);

CREATE INDEX idx_loans_member_id ON loans(member_id);
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_status ON loans(status);

-- migrations/postgres/000004_create_loans_table.down.sql
DROP TABLE IF EXISTS loans;
```

### Quick Database Reset

```bash
# Nuclear option - destroys all data
make down
docker volume rm $(docker volume ls -q | grep library)
make up
make migrate-up
```

## Testing Recipes

### Run Single Test

```bash
# Run specific test by name
go test -v -run TestCreateBook ./internal/usecase/bookops/

# Run specific test in specific file
go test -v -run TestService_ValidateISBN ./internal/domain/book/

# Run with coverage
go test -v -run TestCreateBook -coverprofile=coverage.out ./internal/usecase/bookops/
go tool cover -html=coverage.out
```

### Test with Race Detection

```bash
# Detect race conditions
go test -race ./...

# Specific package
go test -race ./internal/usecase/bookops/
```

### Integration Test Recipe

```go
//go:build integration

package integration_test

import (
    "context"
    "testing"

    "library-service/internal/adapters/repository/postgres"
    "library-service/internal/domain/book"
    "library-service/test/testdb"
)

func TestBookRepository_Integration(t *testing.T) {
    // Setup test database
    db := testdb.Setup(t)
    defer testdb.Teardown(t, db)

    // Create repository
    repo := postgres.NewBookRepository(db)
    ctx := context.Background()

    // Test create
    newBook := book.NewEntity("Test Book", "9780132350884", "Technology")
    err := repo.Create(ctx, newBook)
    if err != nil {
        t.Fatalf("Failed to create book: %v", err)
    }

    // Test retrieve
    retrieved, err := repo.GetByID(ctx, newBook.ID)
    if err != nil {
        t.Fatalf("Failed to get book: %v", err)
    }

    if retrieved.Name != newBook.Name {
        t.Errorf("Got name %s, want %s", retrieved.Name, newBook.Name)
    }
}
```

## API Testing Recipes

### Get JWT Token

```bash
# Register and get token in one command
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}' \
  | jq -r '.tokens.access_token')

echo $TOKEN

# Or login to get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')
```

### Test Protected Endpoints

```bash
# List books
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/books | jq

# Create book
curl -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Clean Code",
    "isbn": "9780132350884",
    "genre": "Technology",
    "authors": ["robert-martin-id"]
  }' | jq

# Get book
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/books/{book-id} | jq

# Update book
curl -X PUT http://localhost:8080/api/v1/books/{book-id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Clean Code - Updated"}' | jq

# Delete book
curl -X DELETE http://localhost:8080/api/v1/books/{book-id} \
  -H "Authorization: Bearer $TOKEN"
```

### Complete API Test Script

```bash
#!/bin/bash
set -e

BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Library API ==="

# 1. Register
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}')

TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.tokens.access_token')
echo "✓ Got token: ${TOKEN:0:20}..."

# 2. Create book
echo "2. Creating book..."
CREATE_RESPONSE=$(curl -s -X POST $BASE_URL/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Book","isbn":"9780132350884","genre":"Tech","authors":["author-id"]}')

BOOK_ID=$(echo $CREATE_RESPONSE | jq -r '.id')
echo "✓ Created book: $BOOK_ID"

# 3. Get book
echo "3. Getting book..."
curl -s -H "Authorization: Bearer $TOKEN" $BASE_URL/books/$BOOK_ID | jq
echo "✓ Retrieved book"

# 4. List books
echo "4. Listing books..."
curl -s -H "Authorization: Bearer $TOKEN" $BASE_URL/books | jq length
echo "✓ Listed books"

# 5. Delete book
echo "5. Deleting book..."
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" $BASE_URL/books/$BOOK_ID
echo "✓ Deleted book"

echo "=== All tests passed! ==="
```

## Development Recipes

### Hot Reload Setup

```bash
# Using Air (recommended)
# Install
go install github.com/cosmtrek/air@latest

# Create .air.toml
air init

# Run with hot reload
air

# Or add to Makefile:
watch:
    air
```

### Debug with Delve

```bash
# Install
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API
dlv debug ./cmd/api -- --port 8080

# Debug specific test
dlv test ./internal/domain/book -- -test.run TestValidateISBN

# Attach to running process
dlv attach $(pgrep library-api)
```

### Code Generation Recipes

#### Generate Mocks

```bash
# Install mockgen
go install github.com/golang/mock/mockgen@latest

# Generate mock for interface
mockgen -source=internal/domain/book/repository.go \
        -destination=internal/domain/book/mocks/repository.go \
        -package=mocks

# Or add go:generate comment in source file:
//go:generate mockgen -destination=mocks/repository.go -package=mocks . Repository

# Then run:
go generate ./...
```

#### Regenerate Swagger

```bash
# Full regeneration
make gen-docs

# Or manually
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal

# Format swagger comments
swag fmt
```

## Git Recipes

### Clean Branch Workflow

```bash
# Start new feature
git checkout main
git pull origin main
git checkout -b feature/add-loans

# Work on feature...
make ci  # Before committing

# Commit
git add .
git commit -m "feat: add loan management system"

# Keep updated with main
git checkout main
git pull origin main
git checkout feature/add-loans
git rebase main

# Push
git push origin feature/add-loans
```

### Fix Mistakes

```bash
# Uncommit last commit (keep changes)
git reset --soft HEAD~1

# Discard all local changes
git checkout .

# Discard specific file
git checkout -- internal/domain/book/service.go

# Stash changes temporarily
git stash
git checkout main
git stash pop
```

## Docker Recipes

### View Logs

```bash
# All services
cd deployments/docker && docker-compose logs -f

# Specific service
cd deployments/docker && docker-compose logs -f postgres

# Last 100 lines
cd deployments/docker && docker-compose logs --tail=100 postgres
```

### Database Access

```bash
# Connect to PostgreSQL
docker exec -it $(docker ps -qf "name=postgres") psql -U library -d library

# Run SQL file
docker exec -i $(docker ps -qf "name=postgres") psql -U library -d library < backup.sql

# Dump database
docker exec $(docker ps -qf "name=postgres") pg_dump -U library library > backup.sql
```

### Clean Docker

```bash
# Stop all
make down

# Remove volumes
docker volume rm $(docker volume ls -q | grep library)

# Clean system
docker system prune -a --volumes
```

## Performance Recipes

### Benchmark Specific Function

```go
// internal/domain/book/service_benchmark_test.go
func BenchmarkService_ValidateISBN(b *testing.B) {
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
# Make changes...
go test -bench=. -benchmem ./... > new.txt
benchstat old.txt new.txt
```

### Profile API

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/usecase/bookops/
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=. ./internal/usecase/bookops/
go tool pprof mem.prof

# Live profiling (add import _ "net/http/pprof" to main.go)
go tool pprof http://localhost:8080/debug/pprof/profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

## Troubleshooting Recipes

### Port Already in Use

```bash
# Find and kill process
lsof -ti:8080 | xargs kill -9

# Or specific signal
lsof -ti:8080 | xargs kill -SIGTERM
```

### Clean Test Cache

```bash
# Clear all test cache
go clean -testcache

# Clear module cache
go clean -modcache

# Clear all caches
go clean -cache -testcache -modcache
```

### Fix Import Issues

```bash
# Tidy modules
go mod tidy

# Verify modules
go mod verify

# Update all dependencies
go get -u ./...
go mod tidy

# Fix imports
goimports -w .
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check PostgreSQL logs
cd deployments/docker && docker-compose logs postgres

# Test connection
PGPASSWORD=library123 psql -h localhost -U library -d library -c "SELECT 1;"

# Restart services
make down && make up
```

## Quick Fixes

```bash
# Everything is broken after git pull
make down && make up && make migrate-up && go mod tidy && make dev

# Linter won't stop complaining
make fmt && go mod tidy && make lint

# Tests are flaky
go clean -testcache && make test-unit

# Can't connect to database
make down && docker volume rm $(docker volume ls -q | grep library) && make up && make migrate-up

# Swagger not updating
make gen-docs && make run

# Import errors
go mod tidy && goimports -w .
```
