# Repository Pattern Migration Guide

**Purpose:** Guide for migrating existing repositories to use generic helpers and BaseRepository pattern

**Related ADRs:**
- ADR 008: Generic Repository Helpers
- ADR 011: BaseRepository Pattern

**When to use this guide:**
- Refactoring existing repositories
- Creating new repositories
- Reducing repository boilerplate

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Migration Strategy](#migration-strategy)
3. [Pattern 1: Generic Helpers](#pattern-1-generic-helpers)
4. [Pattern 2: BaseRepository](#pattern-2-baserepository)
5. [Step-by-Step Migration](#step-by-step-migration)
6. [Common Patterns](#common-patterns)
7. [Testing Strategy](#testing-strategy)
8. [Troubleshooting](#troubleshooting)

---

## Quick Reference

### Before (Traditional Repository)
```go
type AuthorRepository struct {
    db *sqlx.DB
}

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    query := `SELECT id, full_name, pseudonym FROM authors WHERE id=$1`
    var a author.Author
    err := r.db.GetContext(ctx, &a, query, id)
    return a, HandleSQLError(err)
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    query := `SELECT id, full_name, pseudonym FROM authors ORDER BY id`
    var authors []author.Author
    err := r.db.SelectContext(ctx, &authors, query)
    return authors, err
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM authors WHERE id=$1 RETURNING id`
    err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
    return HandleSQLError(err)
}

// ... 5 more methods (Exists, Count, etc.)
// Total: ~50 lines of boilerplate
```

### After (BaseRepository Pattern)
```go
type AuthorRepository struct {
    postgres.BaseRepository[author.Author]
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: postgres.NewBaseRepository[author.Author](db, "authors"),
    }
}

// ✅ Get, List, Delete, Exists, Count, Transaction - all inherited
// Total: ~7 lines
```

**Code Reduction:** 86% (50 lines → 7 lines)

---

## Migration Strategy

### Decision Tree: Which Pattern to Use?

```
Is this a NEW repository?
├─ YES → Use BaseRepository (Pattern 2)
│         Fast, minimal code, best practices built-in
│
└─ NO → Is this an EXISTING repository with complex queries?
    ├─ YES → Use Generic Helpers (Pattern 1)
    │         Gradual migration, keep custom logic
    │
    └─ NO → Use BaseRepository (Pattern 2)
              Fastest migration, maximum code reduction
```

### Migration Phases

**Phase 1: Low-hanging fruit (1 hour)**
- Migrate simple repositories (Author, BookAuthor)
- Focus on repositories with <200 lines

**Phase 2: Medium complexity (2 hours)**
- Migrate repositories with some custom queries (Book, Member)
- Keep custom methods, inherit standard CRUD

**Phase 3: Complex repositories (3 hours)**
- Migrate repositories with many joins (Reservation, Payment)
- Use BaseRepository for base operations, custom methods for business logic

**Total estimated time:** 6 hours for all 10 repositories

---

## Pattern 1: Generic Helpers

**When to use:**
- ✅ Want gradual migration
- ✅ Need explicit control over queries
- ✅ Repository has many custom methods
- ✅ Want to see the SQL being executed

### Step-by-Step

#### Step 1: Import Generic Package
```go
import (
    "github.com/jmoiron/sqlx"
    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/internal/domain/author"
)
```

#### Step 2: Replace GetByID
```go
// Before
func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    query := `SELECT id, full_name, pseudonym FROM authors WHERE id=$1`
    var a author.Author
    err := r.db.GetContext(ctx, &a, query, id)
    return a, postgres.HandleSQLError(err)
}

// After
func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return postgres.GetByIDWithColumns[author.Author](
        ctx, r.db, "authors",
        "id, full_name, pseudonym",  // Columns
        id,
    )
}
```

#### Step 3: Replace List
```go
// Before
func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    query := `SELECT id, full_name, pseudonym FROM authors ORDER BY id`
    var authors []author.Author
    err := r.db.SelectContext(ctx, &authors, query)
    return authors, err
}

// After
func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    return postgres.ListWithColumns[author.Author](
        ctx, r.db, "authors",
        "id, full_name, pseudonym",  // Columns
        "id",                        // Order by
    )
}
```

#### Step 4: Replace Delete
```go
// Before
func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM authors WHERE id=$1 RETURNING id`
    err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
    return postgres.HandleSQLError(err)
}

// After
func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    return postgres.DeleteByID(ctx, r.db, "authors", id)
}
```

#### Step 5: Add Utility Methods
```go
// Add Exists method
func (r *AuthorRepository) Exists(ctx context.Context, id string) (bool, error) {
    return postgres.ExistsByID(ctx, r.db, "authors", id)
}

// Add Count method
func (r *AuthorRepository) Count(ctx context.Context) (int64, error) {
    return postgres.CountAll(ctx, r.db, "authors")
}
```

### Available Generic Helpers

| Helper | Purpose | Example |
|--------|---------|---------|
| `GetByID[T]` | Fetch single entity | `GetByID[Book](ctx, db, "books", id)` |
| `GetByIDWithColumns[T]` | Fetch with specific columns | `GetByIDWithColumns[Book](ctx, db, "books", "id, title", id)` |
| `List[T]` | Fetch all entities | `List[Book](ctx, db, "books", "title")` |
| `ListWithColumns[T]` | Fetch with columns + order | `ListWithColumns[Book](ctx, db, "books", "id, title", "created_at DESC")` |
| `DeleteByID` | Delete by ID | `DeleteByID(ctx, db, "books", id)` |
| `ExistsByID` | Check existence | `ExistsByID(ctx, db, "books", id)` |
| `CountAll` | Count all entities | `CountAll(ctx, db, "books")` |

---

## Pattern 2: BaseRepository

**When to use:**
- ✅ Creating new repository
- ✅ Want maximum code reduction
- ✅ Standard CRUD operations sufficient
- ✅ Happy with defaults (SELECT *, order by ID)

### Step-by-Step

#### Step 1: Define Repository with Embedded Base
```go
package postgres

import (
    "context"
    "github.com/jmoiron/sqlx"
    "library-service/internal/domain/author"
)

type AuthorRepository struct {
    postgres.BaseRepository[author.Author]  // Embed generic base
}
```

#### Step 2: Update Constructor
```go
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: postgres.NewBaseRepository[author.Author](db, "authors"),
    }
}
```

#### Step 3: Remove Inherited Methods

**Delete these methods (now inherited):**
- ❌ `GetByID()` - Use `repo.Get(ctx, id)`
- ❌ `List()` - Use `repo.List(ctx)`
- ❌ `Delete()` - Use `repo.Delete(ctx, id)`
- ❌ `Exists()` - Use `repo.Exists(ctx, id)`
- ❌ `Count()` - Use `repo.Count(ctx)`

#### Step 4: Implement Entity-Specific Methods

**Keep these (Add/Update require column mapping):**
```go
func (r *AuthorRepository) Add(ctx context.Context, a author.Author) (string, error) {
    id := r.GenerateID()  // ✅ Use inherited utility
    query := `INSERT INTO authors (id, full_name, pseudonym, specialty)
              VALUES ($1, $2, $3, $4)`
    _, err := r.GetDB().ExecContext(ctx, query, id, a.FullName, a.Pseudonym, a.Specialty)
    return id, postgres.HandleSQLError(err)
}

func (r *AuthorRepository) Update(ctx context.Context, id string, a author.Author) error {
    query := `UPDATE authors
              SET full_name=$1, pseudonym=$2, specialty=$3
              WHERE id=$4`
    _, err := r.GetDB().ExecContext(ctx, query, a.FullName, a.Pseudonym, a.Specialty, id)
    return postgres.HandleSQLError(err)
}
```

#### Step 5: Custom Business Methods
```go
// Custom query (not covered by base repository)
func (r *AuthorRepository) FindByPseudonym(ctx context.Context, pseudonym string) ([]author.Author, error) {
    query := `SELECT * FROM authors WHERE pseudonym ILIKE $1`
    var authors []author.Author
    err := r.GetDB().SelectContext(ctx, &authors, query, "%"+pseudonym+"%")
    return authors, postgres.HandleSQLError(err)
}
```

### BaseRepository Inherited Methods

| Method | Signature | Purpose |
|--------|-----------|---------|
| `Get` | `(ctx, id) (T, error)` | Retrieve single entity |
| `List` | `(ctx) ([]T, error)` | List all (ordered by ID) |
| `ListWithOrder` | `(ctx, orderBy) ([]T, error)` | List with custom ordering |
| `Delete` | `(ctx, id) error` | Delete by ID |
| `Exists` | `(ctx, id) (bool, error)` | Check existence |
| `Count` | `(ctx) (int64, error)` | Count entities |
| `BatchGet` | `(ctx, ids []string) ([]T, error)` | Fetch multiple by IDs |
| `GenerateID` | `() string` | Generate UUID |
| `GetDB` | `() *sqlx.DB` | Access database connection |
| `GetTableName` | `() string` | Get table name |
| `Transaction` | `(ctx, fn) error` | Execute in transaction |

---

## Step-by-Step Migration

### Example: Migrating BookRepository

#### Current State (book.go)
```go
package postgres

import (
    "context"
    "github.com/jmoiron/sqlx"
    "library-service/internal/domain/book"
)

type BookRepository struct {
    db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
    return &BookRepository{db: db}
}

func (r *BookRepository) GetByID(ctx context.Context, id string) (book.Book, error) {
    query := `SELECT * FROM books WHERE id=$1`
    var b book.Book
    err := r.db.GetContext(ctx, &b, query, id)
    return b, HandleSQLError(err)
}

func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) {
    query := `SELECT * FROM books ORDER BY title`
    var books []book.Book
    err := r.db.SelectContext(ctx, &books, query)
    return books, err
}

func (r *BookRepository) Add(ctx context.Context, b book.Book) (string, error) {
    id := uuid.New().String()
    query := `INSERT INTO books (...) VALUES (...)`
    _, err := r.db.ExecContext(ctx, query, id, ...)
    return id, HandleSQLError(err)
}

func (r *BookRepository) Update(ctx context.Context, id string, b book.Book) error {
    query := `UPDATE books SET ... WHERE id=$1`
    _, err := r.db.ExecContext(ctx, query, ..., id)
    return HandleSQLError(err)
}

func (r *BookRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM books WHERE id=$1 RETURNING id`
    err := r.db.QueryRowContext(ctx, query, id).Scan(&id)
    return HandleSQLError(err)
}

func (r *BookRepository) FindByISBN(ctx context.Context, isbn string) (book.Book, error) {
    query := `SELECT * FROM books WHERE isbn=$1`
    var b book.Book
    err := r.db.GetContext(ctx, &b, query, isbn)
    return b, HandleSQLError(err)
}

// ... more methods
```

#### Step 1: Embed BaseRepository
```go
type BookRepository struct {
    db *sqlx.DB  // Keep temporarily
    BaseRepository[book.Book]  // Add embedded base
}
```

#### Step 2: Update Constructor
```go
func NewBookRepository(db *sqlx.DB) *BookRepository {
    return &BookRepository{
        db: db,  // Keep temporarily
        BaseRepository: NewBaseRepository[book.Book](db, "books"),
    }
}
```

#### Step 3: Delete Standard CRUD Methods

**Delete these (now inherited):**
```go
// ❌ DELETE
// func (r *BookRepository) GetByID(ctx context.Context, id string) (book.Book, error) { ... }
// func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) { ... }
// func (r *BookRepository) Delete(ctx context.Context, id string) error { ... }
```

#### Step 4: Override List with Custom Ordering
```go
// Keep if you want custom ordering
func (r *BookRepository) List(ctx context.Context) ([]book.Book, error) {
    return r.ListWithOrder(ctx, "title ASC")  // ✅ Use inherited method
}
```

#### Step 5: Update Add/Update to Use Utilities
```go
func (r *BookRepository) Add(ctx context.Context, b book.Book) (string, error) {
    id := r.GenerateID()  // ✅ Use inherited method instead of uuid.New()
    query := `INSERT INTO books (id, title, isbn, ...) VALUES ($1, $2, $3, ...)`
    _, err := r.GetDB().ExecContext(ctx, query, id, b.Title, b.ISBN, ...)
    return id, HandleSQLError(err)
}

func (r *BookRepository) Update(ctx context.Context, id string, b book.Book) error {
    query := `UPDATE books SET title=$1, isbn=$2, ... WHERE id=$10`
    _, err := r.GetDB().ExecContext(ctx, query, b.Title, b.ISBN, ..., id)
    return HandleSQLError(err)
}
```

#### Step 6: Keep Custom Business Methods
```go
// ✅ Keep custom methods
func (r *BookRepository) FindByISBN(ctx context.Context, isbn string) (book.Book, error) {
    query := `SELECT * FROM books WHERE isbn=$1`
    var b book.Book
    err := r.GetDB().ExecContext(ctx, &b, query, isbn)
    return b, HandleSQLError(err)
}
```

#### Step 7: Run Tests
```bash
go test ./internal/infrastructure/pkg/repository/postgres -run TestBookRepository
```

#### Step 8: Remove Temporary `db` Field
```go
type BookRepository struct {
    BaseRepository[book.Book]  // ✅ Only this remains
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
    return &BookRepository{
        BaseRepository: NewBaseRepository[book.Book](db, "books"),
    }
}
```

---

## Common Patterns

### Pattern: Transaction Support
```go
func (r *BookRepository) AddWithAuthors(ctx context.Context, b book.Book, authorIDs []string) error {
    return r.Transaction(ctx, func(tx *sqlx.Tx) error {
        // Insert book
        _, err := tx.ExecContext(ctx, `INSERT INTO books ...`, ...)
        if err != nil {
            return err
        }

        // Insert book-author relationships
        for _, authorID := range authorIDs {
            _, err := tx.ExecContext(ctx, `INSERT INTO book_authors ...`, ...)
            if err != nil {
                return err
            }
        }

        return nil
    })
}
```

### Pattern: Batch Operations
```go
func (r *BookRepository) GetBooksForOrder(ctx context.Context, bookIDs []string) ([]book.Book, error) {
    return r.BatchGet(ctx, bookIDs)  // ✅ Uses PostgreSQL ANY($1)
}
```

### Pattern: Custom Query with Inherited Utilities
```go
func (r *MemberRepository) FindActiveMembers(ctx context.Context) ([]member.Member, error) {
    // Check if table has members first
    count, err := r.Count(ctx)  // ✅ Use inherited Count
    if err != nil {
        return nil, err
    }
    if count == 0 {
        return []member.Member{}, nil
    }

    // Custom query
    query := `SELECT * FROM members WHERE active=true ORDER BY created_at DESC`
    var members []member.Member
    err = r.GetDB().SelectContext(ctx, &members, query)
    return members, HandleSQLError(err)
}
```

---

## Testing Strategy

### Unit Testing BaseRepository

**BaseRepository is already tested** (`base_test.go` - 9 test functions).

**You DON'T need to test inherited methods:**
```go
// ❌ DELETE - No longer needed
func TestAuthorRepository_GetByID(t *testing.T) { ... }
func TestAuthorRepository_List(t *testing.T) { ... }
func TestAuthorRepository_Delete(t *testing.T) { ... }
```

**You SHOULD test custom methods:**
```go
// ✅ KEEP - Test entity-specific logic
func TestAuthorRepository_Add(t *testing.T) { ... }
func TestAuthorRepository_Update(t *testing.T) { ... }
func TestAuthorRepository_FindByPseudonym(t *testing.T) { ... }
```

### Integration Testing

**Create integration tests for:**
- Add/Update operations (entity-specific SQL)
- Custom business queries
- Transaction scenarios

```go
//go:build integration

func TestAuthorRepository_Integration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    repo := NewAuthorRepository(db)

    t.Run("Add and Get author", func(t *testing.T) {
        a := author.Author{
            FullName:  "Jane Doe",
            Pseudonym: strPtr("J.D."),
            Specialty: "Science Fiction",
        }

        id, err := repo.Add(ctx, a)
        require.NoError(t, err)

        retrieved, err := repo.Get(ctx, id)  // ✅ Uses inherited Get
        require.NoError(t, err)
        assert.Equal(t, a.FullName, retrieved.FullName)
    })
}
```

---

## Troubleshooting

### Issue: "Cannot use BaseRepository[T]"
**Error:**
```
cannot use BaseRepository[author.Author] (type BaseRepository[author.Author])
as type *BaseRepository[author.Author] in field value
```

**Solution:**
```go
// ❌ WRONG
type AuthorRepository struct {
    *BaseRepository[author.Author]  // Don't use pointer
}

// ✅ CORRECT
type AuthorRepository struct {
    BaseRepository[author.Author]  // Embed by value
}
```

### Issue: "Method requires receiver of type *BaseRepository"
**Error:**
```
repo.Get undefined (type AuthorRepository has no field or method Get)
```

**Solution:**
Ensure you're using pointer receiver:
```go
// ✅ CORRECT
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {  // Return pointer
    return &AuthorRepository{
        BaseRepository: NewBaseRepository[author.Author](db, "authors"),
    }
}
```

### Issue: "SELECT * not working for my entity"
**Problem:** Entity has unexported fields or different column names

**Solution:** Use `GetByIDWithColumns` instead:
```go
func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return GetByIDWithColumns[author.Author](
        ctx, r.db, "authors",
        "id, full_name, pseudonym, specialty",  // Explicit columns
        id,
    )
}
```

### Issue: "Tests fail after migration"
**Problem:** Tests rely on specific SQL queries

**Solution:** Update tests to verify behavior, not SQL:
```go
// ❌ BAD - Fragile SQL assertion
mock.ExpectQuery("SELECT id, name FROM authors WHERE id=$1")

// ✅ GOOD - Verify result
author, err := repo.Get(ctx, id)
assert.NoError(t, err)
assert.Equal(t, "Expected Name", author.FullName)
```

---

## Checklist

### Before Migration
- [ ] Read ADR 008 (Generic Helpers)
- [ ] Read ADR 011 (BaseRepository)
- [ ] Choose migration pattern (Helpers vs BaseRepository)
- [ ] Review existing repository for custom queries

### During Migration
- [ ] Embed BaseRepository or import generic helpers
- [ ] Update constructor
- [ ] Remove/replace standard CRUD methods
- [ ] Keep entity-specific Add/Update methods
- [ ] Keep custom business query methods
- [ ] Run tests after each change

### After Migration
- [ ] All tests passing
- [ ] Code review for consistency
- [ ] Update repository documentation
- [ ] Remove unused imports (uuid, etc.)

---

## Summary

**Quick Migration:**
1. **New repository:** Start with BaseRepository pattern
2. **Existing repository:** Use generic helpers for gradual migration
3. **Complex repository:** BaseRepository + custom methods

**Code Reduction:**
- Generic Helpers: ~80% reduction for standard methods
- BaseRepository: ~86% reduction for entire repository

**Testing:**
- BaseRepository methods already tested
- Focus on testing Add/Update and custom methods
- Integration tests for full workflows

**Next Steps:**
- See `.claude/adrs/` for detailed architecture decisions
- Review `generic_test.go` and `base_test.go` for examples
- Migrate repositories one at a time, test thoroughly
