# Cache Implementation Migration Plan

**Date:** October 11, 2025
**Goal:** Move cache implementations to bounded contexts for better cohesion

---

## Current State Analysis

### Current Structure
```
internal/infrastructure/pkg/cache/
├── cache.go                    # Container coordinating all caches
├── warming.go                  # Cache warming (cross-context)
├── warming_test.go
├── memory/
│   ├── author.go              # Memory cache for authors
│   └── book.go                # Memory cache for books
└── redis/
    ├── author.go              # Redis cache for authors
    └── book.go                # Redis cache for books
```

### Issues
1. **Cache implementations separated from their domains** - Book/Author cache implementations are in shared adapters, not in the books bounded context
2. **Violates bounded context principles** - Books context not fully self-contained
3. **Mixed concerns** - General infrastructure (warming, container) mixed with domain-specific implementations

---

## Proposed Structure

### Target Structure
```
internal/books/cache/              # ✅ NEW - Books context owns its cache
├── memory/
│   ├── book.go                    # Book memory cache
│   └── author.go                  # Author memory cache
└── redis/
    ├── book.go                    # Book redis cache
    └── author.go                  # Author redis cache

internal/infrastructure/pkg/cache/           # Infrastructure coordination only
├── cache.go                       # Container (coordinates all caches)
├── warming.go                     # Cache warming (cross-context)
└── warming_test.go
```

---

## Migration Steps

### Step 1: Create Books Cache Directory ✅
```bash
mkdir -p internal/books/cache/memory
mkdir -p internal/books/cache/redis
```

### Step 2: Move Cache Implementations ✅
```bash
# Move memory implementations
mv internal/infrastructure/pkg/cache/memory/book.go internal/books/cache/memory/
mv internal/infrastructure/pkg/cache/memory/author.go internal/books/cache/memory/

# Move redis implementations
mv internal/infrastructure/pkg/cache/redis/book.go internal/books/cache/redis/
mv internal/infrastructure/pkg/cache/redis/author.go internal/books/cache/redis/
```

### Step 3: Update Package Names ✅
Change package declarations:
- `package memory` → remains `memory`
- `package redis` → remains `redis`

### Step 4: Update Imports ✅
In `internal/infrastructure/pkg/cache/cache.go`:
```go
// OLD
import (
    "library-service/internal/infrastructure/pkg/cache/memory"
    "library-service/internal/infrastructure/pkg/cache/redis"
)

// NEW
import (
    "library-service/internal/books/cache/memory"
    "library-service/internal/books/cache/redis"
)
```

### Step 5: Cleanup Old Directories ✅
```bash
# Remove empty directories
rmdir internal/infrastructure/pkg/cache/memory
rmdir internal/infrastructure/pkg/cache/redis
```

### Step 6: Add Documentation ✅
Create `internal/books/cache/doc.go`:
```go
// Package cache provides cache implementations for the books bounded context.
//
// This package contains both memory and Redis cache implementations for
// Book and Author entities, keeping cache infrastructure colocated with
// the domain it serves.
package cache
```

---

## Benefits

1. **Bounded Context Cohesion** - Books context now contains all its infrastructure (domain, operations, http, repository, cache)
2. **Better Organization** - Cache implementations are next to what they cache
3. **Clearer Separation** - Shared infrastructure (warming, container) vs domain-specific (implementations)
4. **Easier to Find** - Developers look in books/ for all book-related code
5. **Scalable Pattern** - Other contexts can have their own cache implementations

---

## Validation

### Before
```
$ find internal/infrastructure/pkg/cache -name "*.go" | wc -l
10
```

### After
```
$ find internal/infrastructure/pkg/cache -name "*.go" | wc -l
3  # Only cache.go, warming.go, warming_test.go

$ find internal/books/cache -name "*.go" | wc -l
5  # memory/book.go, memory/author.go, redis/book.go, redis/author.go, doc.go
```

---

## Risk Assessment

**Risk Level:** LOW

- **No behavior changes** - Just moving files and updating imports
- **Type-safe** - Compiler will catch any missed imports
- **Fully tested** - Cache warming tests verify functionality
- **Clear rollback** - Git revert if issues arise

---

## Implementation

Execute migration and verify all tests pass.
