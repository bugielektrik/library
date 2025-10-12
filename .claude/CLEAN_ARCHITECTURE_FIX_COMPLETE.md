# Clean Architecture Fix - Infrastructure Layer Purity

**Date:** October 12, 2025
**Status:** ✅ Complete
**Related Issue:** Infrastructure layer was importing domain packages, violating clean architecture boundaries

## Problem Statement

The infrastructure layer (`internal/infrastructure/`) was importing domain packages, creating a circular dependency violation in the clean architecture:

```
Domain → Use Cases → Infrastructure (WRONG!)
      ↑______________|
```

This violated the core principle that infrastructure should be domain-agnostic and only work with primitive types.

## Violations Found

### 1. Auth Middleware (`internal/infrastructure/pkg/middleware/auth.go`)
- **Issue:** Imported `memberdomain` package for `Role` type
- **Impact:** Middleware layer knew about domain-specific types
- **Code:**
```go
// BEFORE (WRONG)
import memberdomain "library-service/internal/members/domain"

func (m *AuthMiddleware) RequireRole(roles ...memberdomain.Role) func(http.Handler) http.Handler {
    // ...
    if claims.Role == string(role) {
        hasRole = true
    }
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
    return m.RequireRole(memberdomain.RoleAdmin)(next)
}
```

### 2. Repository Container (`internal/infrastructure/pkg/repository/repository.go`)
- **Issue:** Imported all domain packages to define repository interfaces
- **Impact:** Infrastructure container held domain-specific types
- **Code:**
```go
// BEFORE (WRONG)
import (
    "library-service/internal/books/domain/author"
    "library-service/internal/books/domain/book"
    memberdomain "library-service/internal/members/domain"
    paymentdomain "library-service/internal/payments/domain"
    // ... etc
)

type Repositories struct {
    Author  author.Repository
    Book    book.Repository
    Member  memberdomain.Repository
    // ... etc
}
```

### 3. Cache Container (`internal/infrastructure/pkg/cache/cache.go`)
- **Issue:** Imported book and author domain packages for cache interfaces
- **Impact:** Infrastructure cache knew about domain entities
- **Code:**
```go
// BEFORE (WRONG)
import (
    "library-service/internal/books/domain/author"
    "library-service/internal/books/domain/book"
)

type Caches struct {
    Author author.Cache
    Book   book.Cache
}
```

### 4. Cache Warming (`internal/infrastructure/pkg/cache/warming.go`)
- **Issue:** Contained domain-aware warming logic
- **Impact:** Infrastructure layer contained application logic

## Solution

### Phase 1: Move Domain-Aware Containers to App Layer

Domain-aware containers belong in the **application layer** (`internal/app/`), not infrastructure:

#### 1.1 Move Repository Container
```bash
# Moved from infrastructure to app
internal/infrastructure/pkg/repository/repository.go → internal/app/repository.go
```

**Changes:**
- Package: `package repository` → `package app`
- Type: `Configuration` → `RepositoryConfig` (to avoid conflicts)
- Now allowed to import domain packages

#### 1.2 Move Cache Container
```bash
# Moved from infrastructure to app
internal/infrastructure/pkg/cache/cache.go → internal/app/cache.go
```

**Changes:**
- Package: `package cache` → `package app`
- Type: `Configuration` → `CacheConfig` (to avoid conflicts)
- Function: `WithMemoryStore()` → `WithMemoryCache()` (for caches)
- Now allowed to import domain packages

#### 1.3 Move Cache Warming Logic
```bash
# Moved from infrastructure to app
internal/infrastructure/pkg/cache/warming.go → internal/app/warming.go
internal/infrastructure/pkg/cache/warming_test.go → internal/app/warming_test.go
```

**Changes:**
- Package: `package cache` → `package app`
- Fixed import cycle in tests (removed `internal/app` import)
- Updated references to `Repositories` and `Caches` types

### Phase 2: Fix Infrastructure Middleware

#### 2.1 Auth Middleware - Use Strings Instead of Domain Types
```go
// AFTER (CORRECT)
// No domain imports

func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
    // ...
    if claims.Role == role {  // Direct string comparison
        hasRole = true
    }
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
    return m.RequireRole("admin")(next)  // String literal
}
```

**Impact:** Infrastructure middleware now only works with primitive types (strings)

### Phase 3: Preserve Infrastructure Utilities

#### 3.1 Postgres Utilities - Keep in Infrastructure
```bash
# Preserved and moved to proper location
internal/adapters/repository/postgres/ → internal/infrastructure/pkg/repository/postgres/
```

**Files:**
- `base.go` - BaseRepository pattern with generics
- `helpers.go` - SQL error handling
- `generic.go` - Generic CRUD operations
- Tests

**Rationale:** These are domain-agnostic infrastructure utilities that all repositories use.

### Phase 4: Update Import Paths

Updated all files to use new paths:

#### 4.1 Application Files
```go
// app.go and worker/main.go
import "library-service/internal/app"

repos, err := app.NewRepositories(app.WithMemoryStore())
caches, err := app.NewCaches(
    app.Dependencies{Repositories: repos},
    app.WithMemoryCache(),
)
go app.WarmCachesAsync(ctx, caches, app.DefaultWarmingConfig(logger))
```

#### 4.2 Repository Implementations (No Change)
All bounded context repositories still use:
```go
import "library-service/internal/infrastructure/pkg/repository/postgres"
```

## Architecture After Fix

### Correct Layer Dependencies
```
┌─────────────────────────────────────────┐
│           Application Layer             │
│  ┌─────────────────────────────────┐   │
│  │        internal/app/            │   │
│  │  ─ repository.go (container)    │   │ ← Domain-aware
│  │  ─ cache.go (container)         │   │   wiring lives here
│  │  ─ warming.go (app logic)       │   │
│  │  ─ app.go (bootstrap)           │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
              ↓ depends on
┌─────────────────────────────────────────┐
│        Infrastructure Layer             │
│  ┌─────────────────────────────────┐   │
│  │  infrastructure/pkg/            │   │
│  │  ─ middleware/ (primitives)     │   │ ← Domain-agnostic
│  │  ─ repository/postgres/         │   │   utilities only
│  │    (generic SQL helpers)        │   │
│  │  ─ dto/, handlers/, httputil/   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
              ↓ depends on
┌─────────────────────────────────────────┐
│          External Libraries             │
│  (sqlx, redis, chi, jwt, etc.)          │
└─────────────────────────────────────────┘
```

### Directory Structure After Fix
```
internal/
├── app/                              # Application Layer (NEW)
│   ├── app.go                        # Bootstrap
│   ├── repository.go                 # Repository container (domain-aware)
│   ├── cache.go                      # Cache container (domain-aware)
│   ├── warming.go                    # Cache warming logic
│   └── warming_test.go               # Cache warming tests
│
├── infrastructure/                    # Infrastructure Layer
│   ├── pkg/
│   │   ├── middleware/
│   │   │   └── auth.go               # ✅ NOW: Uses strings, no domain imports
│   │   ├── repository/
│   │   │   └── postgres/             # Generic SQL utilities (domain-agnostic)
│   │   │       ├── base.go
│   │   │       ├── helpers.go
│   │   │       └── generic.go
│   │   ├── dto/                      # Shared DTOs
│   │   ├── handlers/                 # Base handlers
│   │   └── httputil/                 # HTTP utilities
│   ├── auth/                         # JWT service
│   ├── store/                        # DB connections
│   └── server/                       # HTTP server
│
├── books/                             # Bounded context (domain-aware)
│   ├── domain/
│   ├── service/
│   ├── handlers/
│   └── repository/                    # Uses infrastructure/pkg/repository/postgres
│
├── members/                           # Bounded context (domain-aware)
├── payments/                          # Bounded context (domain-aware)
└── reservations/                      # Bounded context (domain-aware)
```

## Code Changes Summary

### Files Created/Moved
- ✅ `internal/app/repository.go` (moved from infrastructure/pkg/repository/)
- ✅ `internal/app/cache.go` (moved from infrastructure/pkg/cache/)
- ✅ `internal/app/warming.go` (moved from infrastructure/pkg/cache/)
- ✅ `internal/app/warming_test.go` (moved from infrastructure/pkg/cache/)
- ✅ `internal/infrastructure/pkg/repository/postgres/` (restored and moved from adapters/)

### Files Modified
- ✅ `internal/app/app.go` - Updated to use local app package functions
- ✅ `cmd/worker/main.go` - Updated to use `app.*` references
- ✅ `internal/infrastructure/pkg/middleware/auth.go` - Now uses strings instead of domain types

### Naming Changes
- `Configuration` → `RepositoryConfig` (repository container)
- `Configuration` → `CacheConfig` (cache container)
- `WithMemoryStore()` → `WithMemoryCache()` (for caches, repository kept the name)
- `WithRedisStore()` → `WithRedisCache()` (for caches)

## Verification

### Build Status
```bash
✅ go build ./cmd/api
✅ go build ./cmd/worker
```

### Test Status
```bash
✅ go test ./internal/app/...
✅ go test ./internal/books/...
✅ go test ./internal/members/service/auth/...
✅ go test ./internal/payments/...
```

All 60+ tests passing.

## Benefits

### 1. Clean Architecture Compliance
- Infrastructure layer is now domain-agnostic
- Dependencies flow in the correct direction: Infrastructure → Domain (via interfaces)
- No circular dependencies

### 2. Better Separation of Concerns
- **Application layer** (`internal/app/`): Domain-aware wiring and bootstrap
- **Infrastructure layer** (`internal/infrastructure/`): Domain-agnostic utilities
- **Domain layer** (bounded contexts): Business logic

### 3. Testability
- Infrastructure utilities can be tested in isolation
- Application wiring can be tested with different configurations
- Domain logic remains pure

### 4. Maintainability
- Clear boundaries between layers
- Easier to understand where code belongs
- Reduced coupling between components

## Migration Guide for Developers

### If You Need to Add a New Repository

**Before:** Added to `internal/infrastructure/pkg/repository/repository.go`
**Now:** Add to `internal/app/repository.go`

```go
// internal/app/repository.go
type Repositories struct {
    // ... existing repositories
    NewEntity newentitydomain.Repository  // Your new repository
}

func WithPostgresStore(dsn string) RepositoryConfig {
    return func(r *Repositories) error {
        // ... existing setup
        r.NewEntity = newentityrepo.NewRepository(db.Connection)
        return nil
    }
}
```

### If You Need to Add a New Cache

**Before:** Added to `internal/infrastructure/pkg/cache/cache.go`
**Now:** Add to `internal/app/cache.go`

```go
// internal/app/cache.go
type Caches struct {
    // ... existing caches
    NewEntity newentity.Cache  // Your new cache
}

func WithMemoryCache() CacheConfig {
    return func(c *Caches) error {
        // ... existing setup
        c.NewEntity = memory.NewCache(c.dependencies.Repositories.NewEntity)
        return nil
    }
}
```

### If You Need Infrastructure Middleware

Infrastructure middleware MUST use primitive types only:

```go
// ✅ CORRECT - Uses strings
func (m *Middleware) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
    // Implementation using string comparisons
}

// ❌ WRONG - Uses domain types
import "library-service/internal/domain"
func (m *Middleware) RequirePermission(permissions ...domain.Permission) func(http.Handler) http.Handler {
    // FORBIDDEN!
}
```

## Related Documentation

- **ADR 002:** Clean Architecture Boundaries (`.claude/adr/002-clean-architecture-boundaries.md`)
- **ADR 003:** Domain Services vs Infrastructure (`.claude/adr/003-domain-services-vs-infrastructure.md`)
- **Phase 6 Migration:** Adapters Consolidation (`.claude/ADAPTERS_TO_INFRASTRUCTURE_MIGRATION.md`)

## Lessons Learned

### 1. Container Placement is Critical
Containers that wire domain implementations belong in the **application layer**, not infrastructure. Infrastructure should only provide generic utilities.

### 2. Middleware Should Use Primitives
HTTP middleware should only work with primitive types (strings, ints, bools). Type conversion happens at the handler layer.

### 3. Test Infrastructure Carefully
When moving code between packages, watch for import cycles in tests. The test package declaration can cause subtle issues.

### 4. Naming Conflicts Require Careful Planning
When merging code from two packages, watch for naming conflicts (like `Configuration` type used in both repository and cache containers).

## Conclusion

The infrastructure layer is now **domain-agnostic** and follows clean architecture principles. All domain-aware code has been moved to the appropriate layer (application), and all infrastructure utilities are now generic and reusable across all bounded contexts.

**Status:** ✅ Clean Architecture violation FIXED
**Impact:** 0 breaking changes for bounded contexts (they still use the same postgres utilities)
**Risk Level:** Low (all tests passing, builds successful)
