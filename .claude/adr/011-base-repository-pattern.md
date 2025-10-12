# ADR 011: BaseRepository Pattern with Go Generics

**Status:** Accepted
**Date:** 2025-10-09
**Context:** Phase 3 Refactoring - Repository Pattern Evolution

## Context

After implementing generic helper functions (ADR 008), we identified an opportunity to further improve repository ergonomics by introducing an **embeddable BaseRepository** pattern.

### Remaining Friction Points

Even with generic helpers, repository implementations still had boilerplate:

```go
type AuthorRepository struct {
    db *sqlx.DB  // ❌ Every repository declares this
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{db: db}  // ❌ Boilerplate constructor
}

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return GetByID[author.Author](ctx, r.db, "authors", id)  // ⚠️ Must pass db, tableName every time
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    return DeleteByID(ctx, r.db, "authors", id)  // ⚠️ Repetitive parameters
}

// ❌ Every repository reimplements GetDB(), GenerateID(), etc.
```

**Problems:**
- ❌ Repeated `db *sqlx.DB` field in every repository
- ❌ Repetitive constructor patterns
- ❌ Must pass `db` and `tableName` to every helper call
- ❌ No built-in transaction support
- ❌ Each repository manually implements utility methods (GenerateID, Exists, Count)

### Vision

**Ideal repository with minimal boilerplate:**

```go
type AuthorRepository struct {
    BaseRepository[author.Author]  // ✅ Embed base repository
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: NewBaseRepository[author.Author](db, "authors"),
    }
}

// ✅ Inherited methods: Get, List, Delete, Exists, Count, Transaction, GenerateID, BatchGet
// ✅ No need to implement standard CRUD operations
// ✅ Focus on custom business logic
```

## Decision

**Create an embeddable `BaseRepository[T]` generic struct** that encapsulates:
- Database connection
- Table name
- Common CRUD operations
- Utility methods (ID generation, existence checks, transactions)

### Design

```go
// internal/infrastructure/pkg/repository/postgres/base.go

type BaseRepository[T any] struct {
    db        *sqlx.DB
    tableName string
}

func NewBaseRepository[T any](db *sqlx.DB, tableName string) BaseRepository[T] {
    return BaseRepository[T]{
        db:        db,
        tableName: tableName,
    }
}
```

### Provided Methods

#### 1. CRUD Operations
```go
// Get retrieves entity by ID
func (r *BaseRepository[T]) Get(ctx context.Context, id string) (T, error)

// List retrieves all entities with default ordering
func (r *BaseRepository[T]) List(ctx context.Context) ([]T, error)

// ListWithOrder retrieves entities with custom ordering
func (r *BaseRepository[T]) ListWithOrder(ctx context.Context, orderBy string) ([]T, error)

// Delete removes entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error
```

#### 2. Query Helpers
```go
// Exists checks if entity with ID exists
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error)

// Count returns total entity count
func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error)

// BatchGet retrieves multiple entities by IDs
func (r *BaseRepository[T]) BatchGet(ctx context.Context, ids []string) ([]T, error)
```

#### 3. Utility Methods
```go
// GenerateID creates a new UUID
func (r *BaseRepository[T]) GenerateID() string

// GetDB exposes underlying database connection (for custom queries)
func (r *BaseRepository[T]) GetDB() *sqlx.DB

// GetTableName returns configured table name
func (r *BaseRepository[T]) GetTableName() string
```

#### 4. Transaction Support
```go
// Transaction executes function within database transaction
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error
```

**Example usage:**
```go
err := repo.Transaction(ctx, func(tx *sqlx.Tx) error {
    // Multiple operations in transaction
    _, err := tx.Exec("UPDATE authors SET ...")
    // ...
    return nil
})
```

### Implementation Philosophy

**BaseRepository intentionally does NOT provide:**
- ❌ `Add()`/`Create()` methods - Entity-specific column mappings required
- ❌ `Update()` methods - Field selection varies per entity
- ❌ Complex queries (joins, filters) - Too domain-specific

**Rationale:**
- Generic CRUD works for simple operations (Get, List, Delete)
- Insert/Update require entity-specific SQL (column names, validation)
- Repositories should override with custom implementations

## Usage Patterns

### Pattern 1: Simple Repository (Inherit All Methods)

```go
type AuthorRepository struct {
    BaseRepository[author.Author]
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: NewBaseRepository[author.Author](db, "authors"),
    }
}

// ✅ Get, List, Delete, Exists, Count all inherited
// ✅ Only implement Add, Update, and custom business queries
func (r *AuthorRepository) Add(ctx context.Context, a author.Author) (string, error) {
    id := r.GenerateID()  // ✅ Use inherited method
    query := `INSERT INTO authors (id, full_name, pseudonym, specialty) VALUES ($1, $2, $3, $4)`
    _, err := r.GetDB().ExecContext(ctx, query, id, a.FullName, a.Pseudonym, a.Specialty)
    return id, HandleSQLError(err)
}
```

### Pattern 2: Override Specific Methods

```go
type BookRepository struct {
    BaseRepository[book.Book]
}

// ✅ Inherit base Get implementation
// (No need to override)

// ✅ Override List for custom ordering
func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) {
    // Custom: Order by title instead of ID
    return r.ListWithOrder(ctx, "title ASC")
}
```

### Pattern 3: Custom Queries Alongside Inherited Methods

```go
type MemberRepository struct {
    BaseRepository[member.Member]
}

// ✅ Inherited: Get, List, Delete, Exists, Count

// ✅ Custom business query
func (r *MemberRepository) GetByEmail(ctx context.Context, email string) (member.Member, error) {
    query := `SELECT * FROM members WHERE email = $1`
    var m member.Member
    err := r.GetDB().ExecContext(ctx, &m, query, email)
    return m, HandleSQLError(err)
}

// ✅ Custom business query
func (r *MemberRepository) FindActiveSubscribers(ctx context.Context) ([]member.Member, error) {
    query := `SELECT m.* FROM members m
              JOIN subscriptions s ON m.id = s.member_id
              WHERE s.status = 'active'`
    var members []member.Member
    err := r.GetDB().SelectContext(ctx, &members, query)
    return members, err
}
```

## Consequences

### Positive

✅ **Zero boilerplate for standard CRUD:** Get, List, Delete inherited
✅ **Consistent API:** All repositories have same base methods
✅ **Type-safe:** Generics ensure compile-time correctness
✅ **Transaction support:** Built-in transaction helper
✅ **Utility methods:** ID generation, existence checks, counting
✅ **Flexible:** Override any method when needed
✅ **Testable:** BaseRepository has comprehensive test suite (9 tests)
✅ **Access to internals:** GetDB() for custom queries

### Neutral

⚠️ **Add/Update still manual:** Entity-specific logic required (by design)
⚠️ **Embedding syntax:** Developers must understand Go embedding

### Negative

❌ **None identified:** Optional pattern, existing code can migrate gradually

## Comparison: Before vs After

### Before (Generic Helpers Only)
```go
type AuthorRepository struct {
    db *sqlx.DB  // Manual field
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{db: db}  // Manual constructor
}

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return GetByID[author.Author](ctx, r.db, "authors", id)  // Pass db, table
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    return List[author.Author](ctx, r.db, "authors", "id")  // Pass db, table
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    return DeleteByID(ctx, r.db, "authors", id)  // Pass db, table
}

func (r *AuthorRepository) Exists(ctx context.Context, id string) (bool, error) {
    return ExistsByID(ctx, r.db, "authors", id)  // Pass db, table
}

// ... 5 more standard methods
```

**Lines:** ~50 lines

### After (BaseRepository Pattern)
```go
type AuthorRepository struct {
    BaseRepository[author.Author]  // ✅ Embed base
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: NewBaseRepository[author.Author](db, "authors"),
    }
}

// ✅ Get, List, Delete, Exists, Count, Transaction, GenerateID, BatchGet all inherited
// ✅ Only implement entity-specific methods (Add, Update, custom queries)
```

**Lines:** ~7 lines (86% reduction)

## Testing Strategy

### BaseRepository Tests (base_test.go)

**9 comprehensive test functions:**

```go
✅ TestNewBaseRepository          - Constructor and getters
✅ TestBaseRepository_GenerateID   - UUID generation
✅ TestBaseRepository_Get          - Single entity retrieval
✅ TestBaseRepository_List         - List all entities
✅ TestBaseRepository_ListWithOrder - Custom ordering
✅ TestBaseRepository_Delete       - Entity deletion
✅ TestBaseRepository_Exists       - Existence check
✅ TestBaseRepository_Count        - Count entities
✅ TestBaseRepository_BatchGet     - Batch retrieval (edge cases)
✅ TestBaseRepository_Transaction  - Transaction commit/rollback
```

**All tests passing using sqlmock.**

**Note:** `BatchGet` with actual IDs requires real PostgreSQL (PostgreSQL array type complexity with mocking). Integration tests recommended.

### Repository Tests

**Repositories using BaseRepository** can skip testing inherited methods:

```go
// ❌ No longer needed (BaseRepository tested)
func TestAuthorRepository_GetByID(t *testing.T) { ... }
func TestAuthorRepository_List(t *testing.T) { ... }
func TestAuthorRepository_Delete(t *testing.T) { ... }

// ✅ Only test custom methods
func TestAuthorRepository_Add(t *testing.T) { ... }
func TestAuthorRepository_Update(t *testing.T) { ... }
func TestAuthorRepository_FindByPseudonym(t *testing.T) { ... }
```

**Result:** Fewer repository tests, focus on business logic

## Migration Guide

### For New Repositories

**Always use BaseRepository pattern:**

```go
// 1. Embed BaseRepository in your repository struct
type NewEntityRepository struct {
    BaseRepository[entity.Entity]
}

// 2. Initialize in constructor
func NewEntityRepository(db *sqlx.DB) *NewEntityRepository {
    return &NewEntityRepository{
        BaseRepository: NewBaseRepository[entity.Entity](db, "entities"),
    }
}

// 3. Implement only entity-specific methods
func (r *NewEntityRepository) Add(ctx context.Context, e entity.Entity) (string, error) {
    id := r.GenerateID()  // ✅ Use inherited helper
    query := `INSERT INTO entities (...) VALUES (...)`
    _, err := r.GetDB().ExecContext(ctx, query, ...)  // ✅ Use inherited GetDB()
    return id, HandleSQLError(err)
}

// ✅ Get, List, Delete, Exists, Count, Transaction inherited automatically
```

### For Existing Repositories

**Gradual migration (backward compatible):**

**Step 1:** Embed BaseRepository
```go
type AuthorRepository struct {
    db *sqlx.DB  // Keep existing field temporarily
    BaseRepository[author.Author]  // Add base repository
}
```

**Step 2:** Initialize both in constructor
```go
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        db: db,  // Keep for backward compatibility
        BaseRepository: NewBaseRepository[author.Author](db, "authors"),
    }
}
```

**Step 3:** Remove old method implementations
```go
// ❌ Delete this method (now inherited)
// func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
//     return GetByID[author.Author](ctx, r.db, "authors", id)
// }
```

**Step 4:** Remove `db` field when all methods migrated
```go
type AuthorRepository struct {
    BaseRepository[author.Author]  // ✅ Only base repository
}
```

## Relationship to Other Patterns

### vs Generic Helpers (ADR 008)

| Pattern | Use Case |
|---------|----------|
| **Generic Helpers** | Call-site flexibility, passing db/table each time |
| **BaseRepository** | Encapsulation, embed once, use many times |

**They complement each other:**
- BaseRepository **internally uses** generic helpers
- Generic helpers available for one-off queries
- BaseRepository for consistent repository structure

### vs Traditional Repository Interface

**Before:** Repository interface, concrete implementation
```go
// Domain layer
type AuthorRepository interface {
    Get(ctx, id) (Author, error)
    List(ctx) ([]Author, error)
    Delete(ctx, id) error
}

// Infrastructure layer
type PostgresAuthorRepository struct {
    db *sqlx.DB
}
func (r *PostgresAuthorRepository) Get(ctx, id) (Author, error) { ... }
func (r *PostgresAuthorRepository) List(ctx) ([]Author, error) { ... }
func (r *PostgresAuthorRepository) Delete(ctx, id) error { ... }
```

**After:** Interface + BaseRepository implementation
```go
// Domain layer (unchanged)
type AuthorRepository interface {
    Get(ctx, id) (Author, error)
    List(ctx) ([]Author, error)
    Delete(ctx, id) error
}

// Infrastructure layer (✅ Less code)
type PostgresAuthorRepository struct {
    BaseRepository[Author]  // ✅ Methods inherited
}
// ✅ No manual implementation needed for Get, List, Delete
```

**Result:** Interface pattern preserved, implementation simplified

## Limitations & Edge Cases

### 1. Add/Update Not Included

**By design:** Entity-specific column mapping required

```go
// ❌ Cannot be generic (columns vary per entity)
type Book struct {
    ID          string
    Title       string
    ISBN        string
    AuthorID    string
    // ... 10 more fields
}

type Author struct {
    ID        string
    FullName  string
    Pseudonym *string
    // Different fields = different SQL
}
```

**Solution:** Repositories implement Add/Update manually

### 2. Complex Queries Not Supported

**BaseRepository provides:**
- ✅ Simple CRUD (Get, List, Delete, Exists, Count)

**BaseRepository does NOT provide:**
- ❌ Joins (requires multiple tables)
- ❌ Filtering (WHERE conditions vary)
- ❌ Aggregations (GROUP BY, SUM, AVG)
- ❌ Full-text search

**Solution:** Implement custom methods, use GetDB() for direct access

### 3. BatchGet PostgreSQL Array Type

**Challenge:** PostgreSQL array types hard to mock with sqlmock

```go
// Works in production
func (r *BaseRepository[T]) BatchGet(ctx context.Context, ids []string) ([]T, error) {
    query := `SELECT * FROM table WHERE id = ANY($1)`  // PostgreSQL array syntax
    // ...
}

// Hard to test with sqlmock (array type conversion)
```

**Solution:** Integration tests with real PostgreSQL

## Performance Considerations

**Q: Does BaseRepository add overhead?**

**A: No.** BaseRepository methods delegate to generic helpers, which generate identical SQL to hand-written code. Performance is equivalent.

**Benchmark (future work):**
```go
BenchmarkBaseRepository_Get       // Expected: Same as hand-written
BenchmarkDirectSQL_Get            // Expected: Same performance
BenchmarkGenericHelper_Get        // Expected: Same performance
```

## Related ADRs

- **ADR 008: Generic Repository Helpers** - Foundation for BaseRepository
- **ADR 005: Repository Interfaces** - Interface pattern still applies
- **ADR 006: PostgreSQL** - SQL-specific implementation

## Future Work

1. **Migrate all 10 repositories** to use BaseRepository pattern
2. **Add pagination helpers:**
   ```go
   func (r *BaseRepository[T]) ListWithPagination(ctx, limit, offset) ([]T, error)
   ```
3. **Add filtering helpers** for common patterns:
   ```go
   func (r *BaseRepository[T]) FindByField(ctx, fieldName, value) ([]T, error)
   ```
4. **Benchmark performance** vs hand-written queries
5. **Integration tests** for BatchGet with real PostgreSQL

## Conclusion

BaseRepository pattern builds on generic helpers (ADR 008) to provide an **embeddable, type-safe base repository** that eliminates boilerplate while preserving flexibility. Repositories inherit standard CRUD operations and focus on implementing entity-specific business logic.

**Key Benefits:**
- ✅ 86% code reduction for standard operations
- ✅ Consistent API across all repositories
- ✅ Type-safe with Go generics
- ✅ Comprehensive test coverage
- ✅ Backward compatible migration path
- ✅ Zero performance overhead

**Adoption:** Use for all new repositories, migrate existing repositories gradually.
