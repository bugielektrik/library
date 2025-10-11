# Performance Guide

> **Performance baselines, profiling, and optimization strategies**

## Purpose

Know what's "normal" performance, how to measure it, and when/how to optimize.

**Remember:** Premature optimization is the root of all evil. Optimize based on data, not assumptions.

---

## 📊 Performance Baselines

### Expected Performance Metrics

These are TARGET metrics for a library with 100K books, 50K members, 200K loans:

| Operation | Target | Acceptable | Slow |
|-----------|--------|------------|------|
| **API Endpoints** ||||
| GET /books/:id | < 10ms | < 50ms | > 100ms |
| GET /books (list 50) | < 30ms | < 100ms | > 200ms |
| POST /books (create) | < 20ms | < 100ms | > 200ms |
| POST /loans (borrow book) | < 50ms | < 150ms | > 300ms |
| **Database** ||||
| Simple SELECT by ID | < 5ms | < 20ms | > 50ms |
| SELECT with JOIN (2 tables) | < 15ms | < 50ms | > 100ms |
| SELECT with JOIN (3+ tables) | < 30ms | < 100ms | > 200ms |
| INSERT single row | < 10ms | < 50ms | > 100ms |
| UPDATE single row | < 10ms | < 50ms | > 100ms |
| **Use Cases** ||||
| CreateBook (with validation) | < 30ms | < 100ms | > 200ms |
| GetBook (from cache) | < 1ms | < 5ms | > 10ms |
| GetBook (cache miss) | < 20ms | < 50ms | > 100ms |
| **Domain Logic** ||||
| ValidateISBN | < 0.1ms | < 1ms | > 5ms |
| CalculateLateFee | < 0.01ms | < 0.1ms | > 1ms |

**Note:** These assume:
- Database on same machine (local dev) or < 5ms network latency
- Connection pool properly configured (not creating new connections per request)
- Proper indexes on frequently queried columns

---

## 🎯 Quick Performance Check

```bash
# 1. Benchmark domain logic (should be FAST - pure computation)
go test ./internal/domain/book/ -bench=. -benchmem

# Expected:
# BenchmarkService_ValidateISBN-8    10000000    120 ns/op    0 B/op    0 allocs/op
# BenchmarkService_CalculateLateFee-8    50000000    25 ns/op    0 B/op    0 allocs/op

# 2. Benchmark repository (database operations)
go test ./internal/adapters/repository/postgres/ -bench=. -benchmem

# Expected (with indexes):
# BenchmarkBookRepository_GetByID-8    10000    150000 ns/op    1200 B/op    25 allocs/op  (0.15ms)
# BenchmarkBookRepository_List-8       5000     300000 ns/op    8000 B/op    100 allocs/op (0.3ms)

# 3. Load test API
ab -n 1000 -c 10 http://localhost:8080/api/v1/books

# Expected:
# Requests per second:    200-500 RPS
# Time per request:       20-50ms (mean)
# Failed requests:        0

# 4. Check connection pool
# Database connections should be reused, not created per request
docker exec postgres_container psql -U library -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'library';"

# Expected: 20-25 connections (matching pool size)
# Bad: 100+ connections (pool exhausted or not configured)
```

---

## 🔍 Profiling Techniques

### CPU Profiling

**When to use:** API responses are slow, but database queries are fast.

```bash
# Profile a benchmark
go test ./internal/usecase/bookops/ -bench=BenchmarkCreateBook -cpuprofile=cpu.prof

# Analyze
go tool pprof cpu.prof
```

**In pprof:**
```
(pprof) top
# Shows top 10 functions by CPU time

(pprof) top -cum
# Shows top 10 by cumulative time (includes children)

(pprof) list CreateBookUseCase.Execute
# Shows line-by-line CPU usage in specific function

(pprof) web
# Opens visual graph in browser (requires graphviz)

(pprof) quit
```

**Example output:**
```
Showing nodes accounting for 450ms, 90% of 500ms total
      flat  flat%   sum%        cum   cum%
     150ms 30.00% 30.00%      200ms 40.00%  bookops.CreateBookUseCase.Execute
     100ms 20.00% 50.00%      100ms 20.00%  book.Service.ValidateISBN
      80ms 16.00% 66.00%       80ms 16.00%  postgres.BookRepository.Create
      60ms 12.00% 78.00%       60ms 12.00%  runtime.mallocgc
```

**Interpretation:**
- `flat`: Time spent in function itself
- `cum`: Time spent in function + functions it calls
- `ValidateISBN` taking 100ms? That's slow for validation (should be < 1ms)

---

### Memory Profiling

**When to use:** High memory usage, frequent GC pauses, memory leaks.

```bash
# Profile memory allocations
go test ./internal/usecase/bookops/ -bench=. -memprofile=mem.prof

# Analyze
go tool pprof mem.prof
```

**In pprof:**
```
(pprof) top
# Shows functions allocating most memory

(pprof) list CreateBookUseCase
# Line-by-line memory allocations

(pprof) alloc_space
# Total memory allocated (default view)

(pprof) inuse_space
# Memory currently in use (for finding leaks)
```

**Example output:**
```
      flat  flat%   sum%        cum   cum%
    8.50MB 42.50% 42.50%     8.50MB 42.50%  book.NewEntity
    4.00MB 20.00% 62.50%     4.00MB 20.00%  json.Marshal
    2.50MB 12.50% 75.00%     2.50MB 12.50%  []book.Entity (slice growth)
```

**Common memory issues:**
```go
// ❌ BAD: Allocates new slice on every append (causes frequent re-allocation)
results := []Book{}
for _, item := range items {  // 10,000 items
    results = append(results, process(item))  // Reallocates ~14 times (2, 4, 8, 16, ..., 16384)
}
// Allocates: 10,000 + 10,000 + ... ≈ 160KB wasted

// ✅ GOOD: Pre-allocate with known capacity
results := make([]Book, 0, len(items))  // Single allocation
for _, item := range items {
    results = append(results, process(item))
}
// Allocates: 10,000 items × size = exactly what's needed
```

---

### Goroutine Profiling

**When to use:** Goroutine leaks, deadlocks, high goroutine count.

```bash
# Profile goroutines
go tool pprof http://localhost:8080/debug/pprof/goroutine

# Or in test
go test ./... -bench=. -blockprofile=block.prof
go tool pprof block.prof
```

**Check goroutine count:**
```bash
# While API is running
curl http://localhost:8080/debug/pprof/goroutine?debug=1

# Should see:
# goroutine count: 15-30 (for idle server)
# Bad: 1000+ goroutines (likely leak)
```

**Common goroutine leak:**
```go
// ❌ LEAK: Goroutine never exits
func (uc *UseCase) Execute() {
    ch := make(chan Result)

    go func() {
        result := doWork()
        ch <- result  // ← Blocks forever if nobody receives
    }()

    // If we return without receiving from ch, goroutine leaks
    return nil
}

// ✅ FIXED: Use buffered channel or ensure receiver
func (uc *UseCase) Execute() {
    ch := make(chan Result, 1)  // Buffer size 1

    go func() {
        result := doWork()
        ch <- result  // Won't block even if nobody receives
    }()

    select {
    case result := <-ch:
        return result
    case <-time.After(5 * time.Second):
        return errors.New("timeout")
    }
}
```

---

## 🚀 Optimization Strategies

### Level 1: Low-Hanging Fruit (Do This First)

#### 1. Add Database Indexes

**Check if index exists:**
```sql
-- In postgres
\d books

-- Should show indexes on:
-- - Primary keys (automatic)
-- - Foreign keys (CREATE INDEX idx_loans_book_id ON loans(book_id))
-- - Frequently queried columns (isbn, email, status)
```

**Add indexes for common queries:**
```sql
-- For: WHERE isbn = ?
CREATE INDEX idx_books_isbn ON books(isbn);

-- For: ORDER BY created_at DESC
CREATE INDEX idx_books_created_at_desc ON books(created_at DESC);

-- For: WHERE status = ? AND member_id = ?
CREATE INDEX idx_loans_status_member ON loans(member_id, status);
```

**Impact:** 10-100x faster queries

---

#### 2. Connection Pooling

**Check pool configuration:**
```go
// internal/infrastructure/store/postgres.go
config.MaxConns = 25          // Max connections
config.MinConns = 5           // Min connections
config.MaxConnLifetime = 1h   // Recycle connections
config.MaxConnIdleTime = 30m  // Close idle connections
```

**Too low:** New connections created frequently (slow)
**Too high:** Database overloaded

**Rule of thumb:**
```
MaxConns = ((core_count * 2) + effective_spindle_count)

For modern systems: 20-50 connections
```

**Impact:** 2-5x faster queries

---

#### 3. Avoid N+1 Queries

**❌ N+1 Problem:**
```go
// BAD: 1 query for books + N queries for authors (N = number of books)
books, _ := bookRepo.List(ctx, 50, 0)  // 1 query

for _, book := range books {
    authors, _ := authorRepo.GetByBookID(ctx, book.ID)  // N queries!
    book.Authors = authors
}
// Total: 51 queries for 50 books
```

**✅ Solution: Eager Loading with JOIN:**
```go
// GOOD: 1 query with JOIN
books, _ := bookRepo.ListWithAuthors(ctx, 50, 0)  // 1 query with JOIN
// Total: 1 query
```

**In repository:**
```sql
SELECT
    b.id, b.title, b.isbn,
    a.id AS author_id, a.name AS author_name
FROM books b
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
WHERE b.id = ANY($1)
```

**Impact:** 10-50x faster for lists

---

### Level 2: Caching

#### When to Cache

**Cache these:**
- ✅ Frequently accessed data (hot data)
- ✅ Rarely changing data
- ✅ Expensive computations
- ✅ External API responses

**Don't cache:**
- ❌ Rapidly changing data
- ❌ User-specific data (unless per-user cache)
- ❌ Large objects (> 1MB)
- ❌ Data that MUST be real-time

#### Caching Strategies

**1. In-Memory Cache (Simple)**
```go
type BookCacheService struct {
    cache map[string]book.Entity
    mu    sync.RWMutex
    ttl   time.Duration
}

func (s *BookCacheService) Get(id string) (*book.Entity, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    if book, ok := s.cache[id]; ok {
        return &book, true
    }
    return nil, false
}

func (s *BookCacheService) Set(id string, book book.Entity) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.cache[id] = book
}
```

**2. Redis Cache (Distributed)**
```go
func (uc *GetBookUseCase) Execute(ctx context.Context, id string) (*book.Entity, error) {
    // Check cache
    cacheKey := fmt.Sprintf("book:%s", id)
    if cached, err := uc.redis.Get(ctx, cacheKey).Result(); err == nil {
        var book book.Entity
        json.Unmarshal([]byte(cached), &book)
        return &book, nil
    }

    // Cache miss - get from database
    book, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache (TTL: 1 hour)
    bookJSON, _ := json.Marshal(book)
    uc.redis.Set(ctx, cacheKey, bookJSON, 1*time.Hour)

    return &book, nil
}
```

**Cache Invalidation:**
```go
// When book is updated/deleted, invalidate cache
func (uc *UpdateBookUseCase) Execute(ctx context.Context, book book.Entity) error {
    // Update database
    if err := uc.repo.Update(ctx, book); err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("book:%s", book.ID)
    uc.redis.Del(ctx, cacheKey)

    return nil
}
```

**Impact:** 10-100x faster for cached data

---

### Level 3: Query Optimization

#### Use EXPLAIN ANALYZE

```sql
-- See actual query execution
EXPLAIN ANALYZE
SELECT * FROM books
WHERE status = 'available'
ORDER BY created_at DESC
LIMIT 50;
```

**Output:**
```
Limit (cost=0.42..4.44 rows=50 width=100) (actual time=0.023..0.156 rows=50 loops=1)
  -> Index Scan using idx_books_created_at on books (cost=0.42..8021.44 rows=100000 width=100)
        Filter: (status = 'available')
        Rows Removed by Filter: 200
Planning Time: 0.234 ms
Execution Time: 0.189 ms
```

**Look for:**
- ✅ "Index Scan" (good)
- ❌ "Seq Scan" (bad for large tables)
- ❌ "Rows Removed by Filter" (high number = inefficient)

#### Optimize Slow Queries

**Bad query:**
```sql
-- Seq Scan (slow)
SELECT * FROM books WHERE LOWER(title) LIKE '%gatsby%';
```

**Optimized:**
```sql
-- Create GIN index for full-text search
CREATE INDEX idx_books_title_gin ON books USING GIN (to_tsvector('english', title));

-- Use full-text search
SELECT * FROM books WHERE to_tsvector('english', title) @@ to_tsquery('gatsby');
```

**Impact:** 100x faster for text search

---

### Level 4: Code-Level Optimization

#### Reduce Allocations

**Use sync.Pool for frequently allocated objects:**
```go
var bookPool = sync.Pool{
    New: func() interface{} {
        return &book.Entity{}
    },
}

func (uc *UseCase) Execute() (*book.Entity, error) {
    // Get from pool
    book := bookPool.Get().(*book.Entity)
    defer bookPool.Put(book)  // Return to pool

    // Use book...
    return book, nil
}
```

**Impact:** 30-50% fewer allocations

---

#### Optimize JSON Marshaling

**Use streaming for large responses:**
```go
// ❌ BAD: Load everything into memory
books, _ := repo.List(ctx, 10000, 0)
json.Marshal(books)  // Allocates large buffer

// ✅ GOOD: Stream results
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
    enc := json.NewEncoder(w)

    books, _ := repo.ListStream(ctx)
    for book := range books {
        enc.Encode(book)  // Stream one at a time
    }
}
```

**Impact:** 50% less memory usage

---

## 📏 Benchmarking Best Practices

### Writing Good Benchmarks

```go
func BenchmarkService_ValidateISBN(b *testing.B) {
    svc := book.NewService()
    isbn := "9780743273565"

    b.ResetTimer()  // ← Reset timer after setup

    for i := 0; i < b.N; i++ {
        _ = svc.ValidateISBN(isbn)
    }
}
```

**Run benchmarks:**
```bash
# Run all benchmarks
go test ./... -bench=. -benchmem

# Run specific benchmark
go test ./internal/domain/book/ -bench=BenchmarkService_ValidateISBN

# Compare before/after
go test ./... -bench=. -benchmem > before.txt
# Make changes
go test ./... -bench=. -benchmem > after.txt
benchstat before.txt after.txt
```

**Interpreting results:**
```
BenchmarkService_ValidateISBN-8    10000000    120 ns/op    0 B/op    0 allocs/op
                                   │          │            │         └─ Allocations per op
                                   │          │            └─ Bytes allocated per op
                                   │          └─ Nanoseconds per operation
                                   └─ Iterations (automatically determined)
```

**Goals:**
- Domain logic: < 1,000 ns/op (< 1 μs)
- Use cases: < 100,000 ns/op (< 0.1 ms)
- Repository: < 10,000,000 ns/op (< 10 ms)

---

## ⚠️ When NOT to Optimize

**Don't optimize if:**
1. **No performance problem:** If response times are acceptable, don't optimize
2. **Diminishing returns:** Going from 10ms → 5ms rarely worth the complexity
3. **Wrong bottleneck:** Profile first, optimize the slowest part
4. **Readable code is more valuable:** Unless performance is critical

**Quote to remember:**
> "Premature optimization is the root of all evil" — Donald Knuth

**Optimize when:**
- Performance doesn't meet requirements (see baselines above)
- Profiling shows clear bottleneck
- User experience is impacted
- Infrastructure costs are high

---

## 📋 Performance Optimization Checklist

**Before optimizing:**
- [ ] Measured current performance (benchmark)
- [ ] Identified bottleneck (profiling)
- [ ] Set target performance metric
- [ ] Considered if optimization is worth complexity

**During optimization:**
- [ ] Changed ONE thing at a time
- [ ] Benchmarked after each change
- [ ] Verified tests still pass

**After optimization:**
- [ ] Achieved target performance
- [ ] Added benchmark to prevent regression
- [ ] Documented why optimization was needed

---

**Last Updated:** 2025-01-19
**See Also:**
- [../common-tasks.md](../common-tasks.md) - Optimization guides
- [debugging-guide.md](./debugging-guide.md) - Profiling techniques
