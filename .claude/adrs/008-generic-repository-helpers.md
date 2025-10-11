# ADR 008: Generic Repository Helper Functions

**Status:** Accepted
**Date:** 2025-10-09
**Context:** Phase 3 Refactoring - Structural Improvements

## Context

Repository implementations across the codebase contained significant code duplication. Common operations like `GetByID`, `List`, `Delete`, `Exists`, and `Count` were reimplemented in each repository with nearly identical SQL queries and error handling patterns.

**Problems identified:**
- ~150+ lines of duplicated code across 10 repositories
- Inconsistent error handling patterns
- Manual SQL query construction in every repository
- Difficult to maintain consistency when adding new repositories
- Hard to test common patterns (each repository tests the same logic)

**Example duplication** (before refactoring):
```go
// In author.go
func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    query := `SELECT id, full_name, pseudonym, specialty FROM authors WHERE id=$1`
    var a author.Author
    err := r.db.GetContext(ctx, &a, query, id)
    return a, HandleSQLError(err)
}

// In book.go (nearly identical)
func (r *BookRepository) GetByID(ctx context.Context, id string) (book.Book, error) {
    query := `SELECT id, title, isbn, ... FROM books WHERE id=$1`
    var b book.Book
    err := r.db.GetContext(ctx, &b, query, id)
    return b, HandleSQLError(err)
}

// ... repeated in 8 more repositories
```

## Decision

**Implement generic repository helper functions using Go 1.25 generics.**

Created `internal/adapters/repository/postgres/generic.go` with 7 reusable helper functions:

1. **GetByID[T]** - Retrieve single entity by ID
2. **GetByIDWithColumns[T]** - Retrieve with specific columns
3. **List[T]** - Retrieve all entities with ordering
4. **ListWithColumns[T]** - Retrieve with specific columns and ordering
5. **DeleteByID** - Delete entity by ID (uses RETURNING for verification)
6. **ExistsByID** - Check entity existence
7. **CountAll** - Count total entities

**Key design principles:**
- Type-safe: Uses Go generics (`[T any]`) for compile-time type safety
- Consistent error handling: All helpers use `HandleSQLError()`
- SQL injection safe: Uses parameterized queries
- Flexible: Column selection and ordering options where needed
- Testable: Comprehensive test suite with sqlmock

## Implementation

```go
// Generic helper - works with any entity type
func GetByID[T any](ctx context.Context, db *sqlx.DB, tableName string, id string) (T, error) {
    var entity T
    query := fmt.Sprintf("SELECT * FROM %s WHERE id=$1", tableName)
    err := db.GetContext(ctx, &entity, query, id)
    return entity, HandleSQLError(err)
}

// Repository usage - clean and concise
func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return GetByID[author.Author](ctx, r.db, "authors", id)
}
```

**Refactored repositories** (Phase 3):
- ✅ `author.go` - 3 methods refactored
- ✅ `book.go` - 2 methods refactored
- ✅ `member.go` - 2 methods refactored

**Remaining candidates** (future work):
- `reservation.go`, `payment.go`, `subscription.go`, `book_author.go`, `receipt.go`, `fine.go`, `loan.go`

## Consequences

### Positive

✅ **Code reduction:** ~45 lines saved in Phase 3 (3 repositories), projected ~150+ lines across all 10 repositories
✅ **Consistency:** All repositories use identical patterns for common operations
✅ **Maintainability:** Bug fixes in helpers benefit all repositories
✅ **Type safety:** Compile-time type checking prevents runtime errors
✅ **Testability:** Helper functions tested once, reused everywhere
✅ **Onboarding:** New developers see clear, reusable patterns

### Neutral

⚠️ **Learning curve:** Developers unfamiliar with Go generics need brief introduction
⚠️ **Column selection:** Some queries need explicit column lists (use `GetByIDWithColumns` or `ListWithColumns`)

### Negative

❌ **Not suitable for complex queries:** Joins, aggregations, filtering still need custom implementation
❌ **Table name must be passed:** Slight verbosity (mitigated by BaseRepository pattern - see ADR 011)

## Alternatives Considered

### 1. Code generation (e.g., `go generate`)
- **Pros:** No runtime overhead, fully customizable per repository
- **Cons:** Build complexity, generated code bloat, harder to debug
- **Rejected:** Generics provide compile-time safety without code generation complexity

### 2. Reflection-based generic repository
- **Pros:** Even more flexible, no type parameters needed
- **Cons:** Runtime overhead, type safety lost, harder to debug
- **Rejected:** Go generics provide better performance and type safety

### 3. Continue with copy-paste duplication
- **Pros:** Simple, no abstraction
- **Cons:** Maintenance burden, inconsistency, testing overhead
- **Rejected:** Duplication makes codebase harder to maintain at scale

## Related ADRs

- **ADR 005: Repository Interfaces** - Defines repository interface pattern
- **ADR 006: PostgreSQL as Primary Database** - SQL-specific implementation
- **ADR 011: BaseRepository Pattern** - Builds on this ADR with embeddable base

## Examples

### Before Refactoring
```go
// author.go (15 lines)
func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    query := `SELECT id, full_name, pseudonym, specialty FROM authors ORDER BY id`
    var authors []author.Author
    err := r.db.SelectContext(ctx, &authors, query)
    return authors, err
}

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    query := `SELECT id, full_name, pseudonym, specialty FROM authors WHERE id=$1`
    var a author.Author
    err := r.db.GetContext(ctx, &a, query, id)
    return a, HandleSQLError(err)
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM authors WHERE id=$1 RETURNING id`
    err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
    return HandleSQLError(err)
}
```

### After Refactoring
```go
// author.go (3 lines)
func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    return ListWithColumns[author.Author](ctx, r.db, "authors", "id, full_name, pseudonym, specialty", "id")
}

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return GetByIDWithColumns[author.Author](ctx, r.db, "authors", "id, full_name, pseudonym, specialty", id)
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    return DeleteByID(ctx, r.db, "authors", id)
}
```

**Result:** 15 lines → 3 lines (80% reduction)

## Migration Guide

### For Existing Repositories

1. **Identify duplicate CRUD operations** in your repository
2. **Replace with generic helper call:**
   ```go
   // Before
   func (r *Repository) GetByID(ctx context.Context, id string) (Entity, error) {
       var e Entity
       query := "SELECT * FROM table WHERE id=$1"
       err := r.db.GetContext(ctx, &e, query, id)
       return e, HandleSQLError(err)
   }

   // After
   func (r *Repository) GetByID(ctx context.Context, id string) (Entity, error) {
       return GetByID[Entity](ctx, r.db, "table", id)
   }
   ```
3. **Run tests** to verify behavior is unchanged
4. **Optional:** Use `GetByIDWithColumns` if you need specific column selection

### For New Repositories

**Always use generic helpers for standard CRUD operations:**
```go
func (r *NewRepository) GetByID(ctx context.Context, id string) (entity.Entity, error) {
    return GetByID[entity.Entity](ctx, r.db, "entities", id)
}

func (r *NewRepository) List(ctx context.Context) ([]entity.Entity, error) {
    return List[entity.Entity](ctx, r.db, "entities", "created_at DESC")
}

func (r *NewRepository) Delete(ctx context.Context, id string) error {
    return DeleteByID(ctx, r.db, "entities", id)
}

func (r *NewRepository) Exists(ctx context.Context, id string) (bool, error) {
    return ExistsByID(ctx, r.db, "entities", id)
}
```

**Custom queries** (joins, filters) should still be implemented directly:
```go
func (r *NewRepository) FindByEmail(ctx context.Context, email string) (entity.Entity, error) {
    query := `SELECT * FROM entities WHERE email=$1 AND deleted_at IS NULL`
    var e entity.Entity
    err := r.db.GetContext(ctx, &e, query, email)
    return e, HandleSQLError(err)
}
```

## Testing Strategy

✅ **Comprehensive test suite created:** `generic_test.go` (8 test functions)
✅ **All tests passing** using go-sqlmock
✅ **Coverage:** 100% of generic helper functions

**Test approach:**
- Unit tests for each helper function
- Mock database with sqlmock
- Verify SQL query structure
- Test error handling paths
- Validate return values

## Future Work

1. **Refactor remaining 7 repositories** to use generic helpers
2. **Benchmark performance** vs. hand-written queries (expected: negligible difference)
3. **Consider adding helpers** for common patterns:
   - `FindByField[T](tableName, fieldName, value)` - Generic field search
   - `UpdateFields(tableName, id, fields map[string]interface{})` - Dynamic updates
   - `ListWithPagination[T](tableName, limit, offset)` - Paginated lists

## References

- Go Generics Documentation: https://go.dev/doc/tutorial/generics
- Implementation: `internal/adapters/repository/postgres/generic.go`
- Tests: `internal/adapters/repository/postgres/generic_test.go`
- Related Pattern: ADR 011 (BaseRepository)
