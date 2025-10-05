# Troubleshooting Guide

> **Solutions to common problems in this codebase**

## Quick Diagnostics

```bash
# Check everything at once
echo "=== System Check ===" && \
echo "Go version: $(go version)" && \
echo "Docker running: $(docker ps >/dev/null 2>&1 && echo 'YES' || echo 'NO')" && \
echo "PostgreSQL: $(docker ps | grep postgres >/dev/null && echo 'RUNNING' || echo 'STOPPED')" && \
echo "Redis: $(docker ps | grep redis >/dev/null && echo 'RUNNING' || echo 'STOPPED')" && \
echo "Port 8080: $(lsof -ti:8080 >/dev/null && echo 'IN USE' || echo 'FREE')"
```

## Build & Compilation Issues

### Error: `cannot find package`

**Symptoms:**
```
cannot find package "library-service/internal/domain/book"
```

**Solutions:**
```bash
# 1. Tidy modules
go mod tidy

# 2. Verify modules
go mod verify

# 3. Clear cache and rebuild
go clean -modcache
go mod download
go build ./...
```

### Error: `import cycle not allowed`

**Symptoms:**
```
import cycle not allowed in test
```

**Cause:** Violation of Clean Architecture - Domain importing from outer layers

**Solution:**
1. Check dependency direction: Domain → Use Case → Adapters → Infrastructure
2. Move interfaces to domain layer
3. Use dependency injection instead of direct imports

**Example Fix:**
```go
// ❌ Bad: Use case importing from adapters
// internal/usecase/bookops/create_book.go
import "library-service/internal/adapters/repository/postgres"  // WRONG!

// ✅ Good: Use case depends on domain interface
// internal/usecase/bookops/create_book.go
import "library-service/internal/domain/book"  // Interface defined here

type CreateBookUseCase struct {
    repo book.Repository  // Interface from domain
}
```

### Error: `CGO_ENABLED` issues

**Symptoms:**
```
cgo: C compiler not found
```

**Solution:**
```bash
# Build without CGO (recommended for this project)
CGO_ENABLED=0 go build ./cmd/api

# Or use Makefile (already configured)
make build-api
```

## Database Issues

### Error: `connection refused`

**Symptoms:**
```
dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solutions:**
```bash
# 1. Check if PostgreSQL is running
docker ps | grep postgres

# 2. Start services
make up

# 3. Check logs
cd deployments/docker && docker-compose logs postgres

# 4. Restart services
make down && make up

# 5. Check connection manually
PGPASSWORD=library123 psql -h localhost -U library -d library -c "SELECT 1;"
```

### Error: `pq: database "library" does not exist`

**Symptoms:**
```
pq: database "library" does not exist
```

**Solutions:**
```bash
# 1. Create database (automatic with docker-compose)
make down && make up

# 2. Or manually create
docker exec -it $(docker ps -qf "name=postgres") \
    psql -U library -c "CREATE DATABASE library;"
```

### Error: `relation "books" does not exist`

**Symptoms:**
```
pq: relation "books" does not exist
```

**Cause:** Migrations not run

**Solution:**
```bash
# Run migrations
make migrate-up

# Check migration status
docker exec -it $(docker ps -qf "name=postgres") \
    psql -U library -d library -c "\dt"
```

### Error: `duplicate key value violates unique constraint`

**Symptoms:**
```
pq: duplicate key value violates unique constraint "books_isbn_key"
```

**Cause:** ISBN already exists in database

**Solutions:**
```bash
# 1. Check existing data
docker exec -it $(docker ps -qf "name=postgres") \
    psql -U library -d library -c "SELECT id, isbn FROM books WHERE isbn='9780132350884';"

# 2. Clean test data
make migrate-down && make migrate-up

# 3. Use different ISBN in test
```

### Migration Errors

**Error: `dirty database`**

**Symptoms:**
```
Dirty database version 3. Fix and force version.
```

**Solution:**
```bash
# 1. Check current version
docker exec -it $(docker ps -qf "name=postgres") \
    psql -U library -d library -c "SELECT * FROM schema_migrations;"

# 2. Force version (use carefully!)
go run cmd/migrate/main.go force <version>

# 3. Or reset completely (destructive!)
make migrate-down
make migrate-up
```

## Testing Issues

### Error: Tests pass locally but fail in CI

**Causes:**
- Race conditions
- Test cache
- Time-dependent tests
- Environment differences

**Solutions:**
```bash
# 1. Test with race detector
go test -race ./...

# 2. Clear test cache
go clean -testcache

# 3. Run multiple times
go test -count=10 ./internal/usecase/bookops/

# 4. Check for time dependencies
grep -r "time.Now()" --include="*_test.go"
```

### Error: `panic: runtime error: invalid memory address`

**Cause:** Nil pointer dereference (usually mock not set up)

**Solution:**
```go
// ✅ Good: Always initialize mocks
func TestCreateBook(t *testing.T) {
    mockRepo := &mocks.MockRepository{}
    mockRepo.CreateFunc = func(ctx context.Context, book book.Entity) error {
        return nil  // Explicitly set behavior
    }

    // Test code...
}
```

### Tests Hang Forever

**Cause:** Blocking on channel or context timeout

**Solutions:**
```bash
# Run with timeout
go test -timeout 30s ./...

# Find the hanging test
go test -v -timeout 30s ./... 2>&1 | grep -B5 "panic: test timed out"
```

## API / HTTP Issues

### Error: `401 Unauthorized`

**Symptoms:**
```json
{"error": {"code": "unauthorized", "message": "unauthorized"}}
```

**Solutions:**
```bash
# 1. Check token format
echo $TOKEN  # Should be a long JWT string

# 2. Get new token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# 3. Check Authorization header format
# ✅ Correct: Bearer <token>
# ❌ Wrong: <token> or JWT <token>
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/books

# 4. Check token expiry
# Decode JWT (copy to jwt.io or use jwt-cli)
```

### Error: `404 Not Found` on valid endpoint

**Causes:**
1. Route not registered
2. Wrong HTTP method
3. Wrong base path

**Solutions:**
```bash
# 1. Check registered routes
grep -r "Route\|Get\|Post\|Put\|Delete" internal/adapters/http/router.go

# 2. Check HTTP method
# ✅ POST /api/v1/books
# ❌ GET /api/v1/books (if only POST defined)

# 3. Verify API is running
curl http://localhost:8080/health
```

### Error: `validation failed`

**Symptoms:**
```json
{"error": {"code": "validation_failed", "message": "..."}}
```

**Solutions:**
```bash
# Check validation tags in DTO
# internal/adapters/http/dto/book.go

# Common issues:
# - Missing required field
# - Invalid format (e.g., email, UUID)
# - Value out of range (min/max)

# Example fix:
{
  "name": "Test Book",           // required
  "isbn": "9780132350884",        // required, valid ISBN
  "genre": "Technology",          // required
  "authors": ["valid-uuid-here"]  // required, array of UUIDs
}
```

## Swagger / API Documentation Issues

### Swagger UI Not Loading

**Symptoms:**
- 404 on `/swagger/index.html`
- Blank page

**Solutions:**
```bash
# 1. Regenerate docs
make gen-docs

# 2. Check files exist
ls -la api/openapi/

# 3. Restart server
pkill library-api
make run

# 4. Clear browser cache
# Or try incognito mode
```

### Swagger Annotations Not Updating

**Cause:** Docs not regenerated after code changes

**Solution:**
```bash
# Always regenerate after handler changes
make gen-docs

# Or use watch mode (if using air)
air -c .air.toml
```

### Error: `ParseComment error`

**Symptoms:**
```
cannot find type definition: dto.BookRequest
```

**Cause:** Missing import or type in Swagger comment

**Solution:**
```go
// ✅ Good: Full package path
// @Param request body dto.CreateBookRequest true "Book details"

// ❌ Bad: Missing package or wrong type name
// @Param request body BookRequest true "Book details"
```

## Performance Issues

### Slow Tests

**Diagnosis:**
```bash
# Identify slow tests
go test -v ./... 2>&1 | grep -E "PASS|FAIL" | grep -E "[0-9]+\.[0-9]+s"

# Profile tests
go test -cpuprofile=cpu.prof ./internal/usecase/bookops/
go tool pprof cpu.prof
```

**Common Causes:**
1. Database queries in unit tests → Use mocks
2. External API calls → Mock them
3. Large test data → Reduce size

### Slow Build Times

**Solutions:**
```bash
# 1. Use build cache
go clean -cache
go build -a ./...

# 2. Parallel builds
go build -p 8 ./...

# 3. Remove unnecessary dependencies
go mod tidy

# 4. Check for large vendored dependencies
du -sh vendor/
```

### High Memory Usage

**Diagnosis:**
```bash
# Memory profile
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Check for leaks
go test -race -run TestMemoryLeak ./...
```

**Common Causes:**
- Unclosed database connections
- Goroutine leaks
- Large in-memory caches

## Git Issues

### Error: `conflicts during merge/rebase`

**Solution:**
```bash
# 1. See conflicts
git status

# 2. Fix conflicts manually in editor
# Look for <<<<<<< ======= >>>>>>>

# 3. After fixing
git add .
git rebase --continue  # if rebasing
git commit             # if merging
```

### Accidentally Committed Secrets

**URGENT Solution:**
```bash
# 1. Remove from latest commit
git rm --cached .env
git commit --amend -m "Remove sensitive file"

# 2. Add to .gitignore
echo ".env" >> .gitignore

# 3. If already pushed, rotate secrets immediately!
# Then force push (careful!)
git push --force origin feature-branch
```

## Docker Issues

### Error: `port is already allocated`

**Symptoms:**
```
Bind for 0.0.0.0:5432 failed: port is already allocated
```

**Solutions:**
```bash
# 1. Stop conflicting container
docker ps | grep 5432
docker stop <container-id>

# 2. Or stop local PostgreSQL
brew services stop postgresql  # macOS
sudo systemctl stop postgresql # Linux

# 3. Change port in docker-compose.yml
ports:
  - "5433:5432"  # Use different host port
```

### Error: `no space left on device`

**Solutions:**
```bash
# 1. Clean Docker
docker system prune -a --volumes

# 2. Remove old images
docker image prune -a

# 3. Remove stopped containers
docker container prune
```

## Environment Issues

### Error: `JWT_SECRET is required`

**Solution:**
```bash
# 1. Copy example env
cp .env.example .env

# 2. Set JWT_SECRET
echo "JWT_SECRET=your-super-secret-key-change-in-production" >> .env

# 3. Source environment
export $(cat .env | xargs)
```

### Wrong Go Version

**Symptoms:**
```
go: cannot find main module
```

**Solution:**
```bash
# Check version
go version  # Should be 1.25+

# Install correct version (macOS)
brew install go@1.25

# Or use gvm
gvm install go1.25
gvm use go1.25
```

## Emergency Procedures

### Complete Reset

```bash
# Nuclear option - destroys ALL local data
echo "⚠️  WARNING: This will destroy all local data!"
echo "Press Ctrl+C to cancel, Enter to continue..."
read

# Stop everything
make down

# Remove volumes
docker volume rm $(docker volume ls -q | grep library) 2>/dev/null || true

# Clean Docker
docker system prune -a --volumes -f

# Clean Go
go clean -cache -testcache -modcache

# Rebuild
make init
make up
make migrate-up
make test
make run
```

### Debug Mode

```bash
# Run with maximum verbosity
APP_MODE=dev LOG_LEVEL=debug go run ./cmd/api

# Enable all debugging
export DEBUG=*
export VERBOSE=true
go run ./cmd/api
```

## Getting Help

When asking for help, provide:

1. **Error message** (full output)
2. **What you tried** (commands you ran)
3. **Environment** (go version, OS)
4. **Logs** (from `docker-compose logs` or application logs)

```bash
# Collect diagnostic info
echo "=== Diagnostic Info ===" > debug.txt
echo "Go version: $(go version)" >> debug.txt
echo "Docker: $(docker --version)" >> debug.txt
echo "OS: $(uname -a)" >> debug.txt
echo "" >> debug.txt
echo "=== Running Containers ===" >> debug.txt
docker ps >> debug.txt
echo "" >> debug.txt
echo "=== Postgres Logs ===" >> debug.txt
cd deployments/docker && docker-compose logs --tail=50 postgres >> ../../debug.txt
cd ../..
cat debug.txt
```
