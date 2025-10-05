# Debugging Guide

> **Advanced debugging techniques for Go Clean Architecture**

## Purpose

Deep dive into debugging strategies beyond basic troubleshooting. This guide shows HOW to debug, not just WHAT to check.

**When to use:**
- Basic troubleshooting ([troubleshooting.md](./troubleshooting.md)) didn't solve your issue
- You need to understand deep system behavior
- Production issues require investigation

---

## 🎯 Debugging Toolbox

### Tools You Should Know

```bash
# Delve debugger
dlv debug ./cmd/api/main.go

# Go race detector
go test -race ./...

# PostgreSQL query analyzer
docker exec -it postgres_container psql -U library -c "EXPLAIN ANALYZE SELECT..."

# HTTP request inspector
curl -v http://localhost:8080/api/v1/books

# JSON pretty printer
curl http://localhost:8080/api/v1/books | jq '.'

# Process inspector
lsof -i:8080
ps aux | grep library-api

# Log following
tail -f /var/log/app.log
docker logs -f container_name
```

---

## 🐛 Scenario 1: "Use Case Returns Wrong Data"

### Symptoms
```
GET /books/123 returns book with ID 456
Test passes, but API returns wrong data
```

### Debugging Steps

#### Step 1: Isolate the Layer (5 minutes)

**Test the use case directly:**
```go
// Quick test in *_test.go
func TestDebug_GetBookUseCase(t *testing.T) {
    // Use real repository (not mock)
    db := setupTestDB(t)
    repo := postgres.NewBookRepository(db)

    // Insert known book
    expectedBook := book.Entity{
        ID:    "book-123",
        Title: "Expected Book",
    }
    repo.Create(context.Background(), expectedBook)

    // Test use case
    uc := bookops.NewGetBookUseCase(repo)
    result, err := uc.Execute(context.Background(), "book-123")

    // Debug output
    t.Logf("Expected ID: %s, Got ID: %s", expectedBook.ID, result.ID)
    assert.Equal(t, "book-123", result.ID)
}
```

**Run:**
```bash
go test ./internal/usecase/bookops/ -v -run TestDebug_GetBookUseCase
```

**If test PASSES:** Bug is in HTTP layer (handler, routing, middleware)
**If test FAILS:** Bug is in use case or repository

---

#### Step 2a: If Bug in HTTP Layer

**Add logging to handler:**
```go
// internal/adapters/http/handlers/book.go
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")

    // DEBUG: Log request
    log.Printf("DEBUG: GetBook called with ID: %s", id)

    book, err := h.getBookUC.Execute(r.Context(), id)

    // DEBUG: Log result
    log.Printf("DEBUG: Use case returned book ID: %s, Title: %s", book.ID, book.Title)

    // ... rest of handler
}
```

**Test with curl:**
```bash
# Make request
curl -v http://localhost:8080/api/v1/books/book-123

# Check logs
tail -f logs/app.log | grep DEBUG
```

**Common issues:**
- Wrong ID extracted from URL (check chi.URLParam)
- Middleware modifying context
- Wrong routing (check router.go)

---

#### Step 2b: If Bug in Use Case/Repository

**Add logging to use case:**
```go
// internal/usecase/bookops/get_book.go
func (uc *GetBookUseCase) Execute(ctx context.Context, id string) (*book.Entity, error) {
    log.Printf("DEBUG: GetBookUseCase.Execute called with ID: %s", id)

    book, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        log.Printf("DEBUG: Repository error: %v", err)
        return nil, err
    }

    log.Printf("DEBUG: Repository returned book ID: %s, Title: %s", book.ID, book.Title)
    return &book, nil
}
```

**Add logging to repository:**
```go
// internal/adapters/repository/postgres/book.go
func (r *BookRepository) GetByID(ctx context.Context, id string) (book.Entity, error) {
    query := "SELECT id, title, isbn FROM books WHERE id = $1"

    log.Printf("DEBUG: Executing query: %s with ID: %s", query, id)

    var b book.Entity
    row := r.db.QueryRowContext(ctx, query, id)
    err := row.Scan(&b.ID, &b.Title, &b.ISBN)

    log.Printf("DEBUG: Query returned ID: %s, Title: %s", b.ID, b.Title)

    return b, err
}
```

**Run and check logs:**
```bash
make run
# In another terminal:
curl http://localhost:8080/api/v1/books/book-123
# Check logs for DEBUG lines
```

**Common issues:**
- SQL query using wrong WHERE clause
- Column mapping wrong in Scan()
- Context timeout/cancellation

---

#### Step 3: Use Delve Debugger

**Start API with debugger:**
```bash
dlv debug ./cmd/api/main.go
```

**In delve:**
```
(dlv) break BookHandler.GetBook
Breakpoint 1 set at 0x... for main.(*BookHandler).GetBook()

(dlv) continue
```

**Make request:**
```bash
curl http://localhost:8080/api/v1/books/book-123
```

**In delve (execution paused at breakpoint):**
```
(dlv) print id
book-123

(dlv) step
# Step through code line by line

(dlv) print book
book.Entity{ID: "book-456", Title: "Wrong Book"}

(dlv) locals
# Print all local variables

(dlv) continue
```

**Delve Commands:**
- `break <function>` - Set breakpoint
- `continue` - Resume execution
- `step` - Step into function
- `next` - Step over function
- `print <var>` - Print variable
- `locals` - Print all local variables
- `stack` - Print stack trace
- `exit` - Exit debugger

---

## 🐛 Scenario 2: "Tests Pass but API Fails"

### Symptoms
```
Unit tests: PASS
Integration tests: PASS
API request: 500 Internal Server Error
```

### Debugging Steps

#### Step 1: Check Logs

```bash
# Application logs
tail -f logs/app.log

# If using Docker
docker logs -f library-api

# Make failing request
curl -v http://localhost:8080/api/v1/books
```

**Look for:**
- Panic stack traces
- Database connection errors
- Missing environment variables

---

#### Step 2: Test with Real Dependencies

**Create integration test:**
```go
// internal/adapters/http/handlers/book_integration_test.go
//go:build integration

func TestIntegration_BookHandler_GetBook(t *testing.T) {
    // Setup REAL dependencies
    cfg := config.Load()
    app, err := app.New(cfg)
    require.NoError(t, err)
    defer app.Close()

    container := container.New(app)
    router := routes.NewRouter(container)

    // Insert test data
    testBook := book.Entity{ID: "book-123", Title: "Test"}
    container.BookRepo.Create(context.Background(), testBook)

    // Make request
    req := httptest.NewRequest("GET", "/api/v1/books/book-123", nil)
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
}
```

**Run:**
```bash
# Start dependencies
make up

# Run integration test
go test ./internal/adapters/http/handlers/ -tags=integration -v
```

**If test FAILS:** Check error message carefully
**If test PASSES:** Issue is in deployment/configuration, not code

---

## 🐛 Scenario 3: "Intermittent Failures / Race Conditions"

### Symptoms
```
Test passes 9/10 times
Test fails randomly
Different results on different runs
```

### Debugging Steps

#### Step 1: Run with Race Detector

```bash
go test -race ./internal/usecase/bookops/ -count=100
```

**If race detected:**
```
WARNING: DATA RACE
Read at 0x00c000124080 by goroutine 23:
  library-service/internal/usecase/bookops.(*CreateBookUseCase).Execute()
      /path/to/create_book.go:45 +0x123

Previous write at 0x00c000124080 by goroutine 19:
  library-service/internal/adapters/repository/postgres.(*BookRepository).Create()
      /path/to/book.go:67 +0x456
```

**Fix:** Use proper synchronization (mutex, channels, atomic)

---

#### Step 2: Add Debugging to Find Race

**Suspicious code:**
```go
// ❌ RACE: shared variable accessed by multiple goroutines
type UseCase struct {
    cache map[string]book.Entity // ← NOT THREAD-SAFE
}

func (uc *UseCase) Execute(id string) (*book.Entity, error) {
    // Goroutine 1 writes
    uc.cache[id] = book

    // Goroutine 2 reads
    if cached, ok := uc.cache[id]; ok {
        return &cached, nil
    }
}
```

**Fix:**
```go
// ✅ FIXED: Use sync.RWMutex
type UseCase struct {
    cache map[string]book.Entity
    mu    sync.RWMutex
}

func (uc *UseCase) Execute(id string) (*book.Entity, error) {
    // Write lock
    uc.mu.Lock()
    uc.cache[id] = book
    uc.mu.Unlock()

    // Read lock
    uc.mu.RLock()
    cached, ok := uc.cache[id]
    uc.mu.RUnlock()

    if ok {
        return &cached, nil
    }
}
```

---

## 🐛 Scenario 4: "SQL Query Returns No Results"

### Symptoms
```
Books exist in database
Query returns empty []
No errors thrown
```

### Debugging Steps

#### Step 1: Log Raw SQL

**Add query logging:**
```go
func (r *BookRepository) List(ctx context.Context, limit, offset int) ([]book.Entity, error) {
    query := "SELECT id, title, isbn FROM books ORDER BY created_at DESC LIMIT $1 OFFSET $2"

    // DEBUG: Log query with parameters
    log.Printf("DEBUG SQL: %s [limit=%d, offset=%d]", query, limit, offset)

    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    // ...
}
```

**Run and check output:**
```bash
make run
curl http://localhost:8080/api/v1/books

# Check logs:
# DEBUG SQL: SELECT id, title... [limit=50, offset=0]
```

---

#### Step 2: Run Query Directly in PostgreSQL

```bash
# Connect to database
docker exec -it $(docker ps -qf "name=postgres") psql -U library -d library

# Run exact query
SELECT id, title, isbn FROM books ORDER BY created_at DESC LIMIT 50 OFFSET 0;

# If returns results: Issue is in Go code (Scan, mapping)
# If returns empty: Issue is in database (no data, wrong table)
```

**Check database:**
```sql
-- Count books
SELECT COUNT(*) FROM books;

-- Check table structure
\d books

-- Check recent data
SELECT * FROM books ORDER BY created_at DESC LIMIT 5;
```

---

#### Step 3: Debug row.Scan()

**Add detailed logging:**
```go
func (r *BookRepository) List(ctx context.Context, limit, offset int) ([]book.Entity, error) {
    query := "SELECT id, title, isbn, status FROM books LIMIT $1 OFFSET $2"

    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var books []book.Entity
    rowCount := 0
    for rows.Next() {
        rowCount++
        var b book.Entity

        err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.Status)
        if err != nil {
            // DEBUG: Log scan error
            log.Printf("DEBUG: Scan error on row %d: %v", rowCount, err)
            return nil, fmt.Errorf("scanning row %d: %w", rowCount, err)
        }

        // DEBUG: Log scanned book
        log.Printf("DEBUG: Scanned book %d: ID=%s, Title=%s", rowCount, b.ID, b.Title)

        books = append(books, b)
    }

    // DEBUG: Log total
    log.Printf("DEBUG: Total rows scanned: %d", rowCount)

    return books, rows.Err()
}
```

**Common Scan() issues:**
```go
// ❌ WRONG: Scan order doesn't match SELECT order
rows.Scan(&b.Title, &b.ID, &b.ISBN) // SELECT id, title, isbn

// ✅ CORRECT: Match SELECT order
rows.Scan(&b.ID, &b.Title, &b.ISBN) // SELECT id, title, isbn

// ❌ WRONG: Wrong number of fields
rows.Scan(&b.ID, &b.Title) // SELECT id, title, isbn (3 fields!)

// ✅ CORRECT: Match number of fields
rows.Scan(&b.ID, &b.Title, &b.ISBN)

// ❌ WRONG: Type mismatch
var count string
rows.Scan(&count) // SELECT COUNT(*) returns int

// ✅ CORRECT: Match types
var count int
rows.Scan(&count)
```

---

## 🐛 Scenario 5: "JWT Token Invalid / Authentication Fails"

### Symptoms
```
Token generated successfully
API returns 401 Unauthorized
Swagger shows "Unauthorized"
```

### Debugging Steps

#### Step 1: Validate Token Structure

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# Print token
echo $TOKEN

# Decode JWT (use jwt.io or command line)
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq '.'
```

**Expected payload:**
```json
{
  "member_id": "uuid",
  "email": "test@example.com",
  "role": "user",
  "iss": "library-service",
  "sub": "uuid",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Check:**
- `exp` (expiration) is in the future
- `iss` (issuer) matches expected value
- All required claims present

---

#### Step 2: Debug Middleware

**Add logging to auth middleware:**
```go
// internal/adapters/http/middleware/auth.go
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")

        // DEBUG: Log header
        log.Printf("DEBUG: Authorization header: %s", authHeader)

        if authHeader == "" {
            log.Printf("DEBUG: No authorization header")
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            log.Printf("DEBUG: Invalid header format: %v", parts)
            http.Error(w, "invalid authorization header", http.StatusUnauthorized)
            return
        }

        token := parts[1]
        log.Printf("DEBUG: Token (first 20 chars): %s...", token[:20])

        claims, err := m.jwtManager.ValidateAccessToken(token)
        if err != nil {
            log.Printf("DEBUG: Token validation error: %v", err)
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        log.Printf("DEBUG: Token valid for member: %s", claims.MemberID)

        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Test:**
```bash
curl -v -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/books

# Check logs for DEBUG lines
```

**Common issues:**
- Missing "Bearer " prefix
- Extra whitespace in header
- Token expired (check `exp` claim)
- Wrong secret used for validation

---

#### Step 3: Check JWT Secret

**In config:**
```bash
# Check environment variable
echo $JWT_SECRET

# Should be same in both places:
# 1. Where token is generated (login use case)
# 2. Where token is validated (middleware)
```

**Add debug to JWT manager:**
```go
// internal/infrastructure/auth/jwt.go
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
    // DEBUG: Log secret (first 10 chars only!)
    log.Printf("DEBUG: Using secret starting with: %s...", m.secret[:10])

    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Check signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(m.secret), nil
    })

    if err != nil {
        log.Printf("DEBUG: Parse error: %v", err)
        return nil, err
    }

    // ...
}
```

---

## 🛠️ Debugging Tools Deep Dive

### Delve Debugger Workflows

**Debug tests:**
```bash
# Debug specific test
dlv test ./internal/usecase/bookops/ -- -test.run TestCreateBook

# In delve:
(dlv) break bookops.CreateBookUseCase.Execute
(dlv) continue
(dlv) print req
(dlv) step
```

**Debug running API:**
```bash
# Attach to running process
ps aux | grep library-api
# Find PID (e.g., 12345)

dlv attach 12345

# Set breakpoints
(dlv) break BookHandler.CreateBook
(dlv) continue
```

**Debug goroutines:**
```bash
(dlv) goroutines
# Lists all goroutines

(dlv) goroutine 5
# Switch to goroutine 5

(dlv) stack
# See stack trace
```

---

### PostgreSQL Query Analysis

**Explain query plans:**
```sql
-- See query plan
EXPLAIN SELECT * FROM books WHERE isbn = '1234567890';

-- See actual execution
EXPLAIN ANALYZE SELECT * FROM books WHERE isbn = '1234567890';
```

**Output:**
```
Seq Scan on books (cost=0.00..1234.56 rows=1 width=100)
  Filter: (isbn = '1234567890'::text)

-- "Seq Scan" = no index, scanning entire table (BAD for large tables)
-- Should use index scan instead
```

**With index:**
```sql
CREATE INDEX idx_books_isbn ON books(isbn);

EXPLAIN ANALYZE SELECT * FROM books WHERE isbn = '1234567890';

-- Output:
Index Scan using idx_books_isbn on books (cost=0.42..8.44 rows=1)
  Index Cond: (isbn = '1234567890'::text)

-- Much better!
```

---

### Memory Profiling

**Profile memory usage:**
```bash
go test ./internal/domain/book/ -memprofile=mem.prof -bench=.

go tool pprof mem.prof

(pprof) top
# Shows functions using most memory

(pprof) list NewEntity
# Shows line-by-line allocations in function
```

**Common memory issues:**
```go
// ❌ BAD: Allocates new slice on every append
for _, item := range items {
    results = append(results, process(item))
}

// ✅ GOOD: Pre-allocate capacity
results := make([]Result, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
```

---

## 📋 Debugging Checklist

Before deep debugging, check these:

- [ ] Tests pass locally (`make test`)
- [ ] Environment variables set correctly
- [ ] Database migrations applied (`make migrate-up`)
- [ ] Dependencies up to date (`go mod tidy`)
- [ ] No port conflicts (`lsof -i:8080`)
- [ ] Logs show no obvious errors

**If all checked and still failing:** Use techniques above

---

**Last Updated:** 2025-01-19
**See Also:** [troubleshooting.md](./troubleshooting.md) for common fixes
