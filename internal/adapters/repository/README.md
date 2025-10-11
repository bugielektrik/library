# Repository Adapter Layer

Database implementations of domain repository interfaces.

## Structure

```
repository/
├── postgres/    # PostgreSQL implementations
│   ├── base.go  # Base repository with common CRUD
│   ├── book.go  # Book repository
│   ├── member.go # Member repository
│   ├── payment.go # Payment repository
│   └── generic.go # Generic CRUD operations
├── memory/      # In-memory implementations (testing)
├── mocks/       # Generated mocks for testing
└── errors.go    # SQL error handling
```

## PostgreSQL Implementation

### Base Repository

Provides common functionality:
- Database connection access
- Transaction support
- Error handling
- Common CRUD operations

```go
type BaseRepository struct {
    db *sqlx.DB
}

func (r *BaseRepository) GetDB() *sqlx.DB {
    return r.db
}
```

### Repository Pattern

Each repository:
1. Implements domain interface
2. Extends BaseRepository
3. Maps between domain entities and DB

```go
type BookRepository struct {
    *BaseRepository
}

func (r *BookRepository) Create(ctx context.Context, book book.Book) (string, error) {
    // SQL implementation
}
```

### Error Handling

`HandleSQLError()` converts database errors to domain errors:
- Duplicate key → `ErrAlreadyExists`
- Not found → `ErrNotFound`
- Foreign key → `ErrValidation`

## Memory Implementation

Used for:
- Unit testing
- Local development
- Caching layer

Features:
- Thread-safe with mutex
- No external dependencies
- Fast operations

## Mocks

Generated using mockgen:
```bash
mockgen -source=domain/book/repository.go -destination=mocks/book_repository_mock.go
```

Used in use case tests to isolate business logic.

## Generic Operations

`generic.go` provides type-safe CRUD:
```go
GetByID[T any](ctx, id, table) (T, error)
List[T any](ctx, table, limit, offset) ([]T, error)
DeleteByID(ctx, id, table) error
```

## Best Practices

1. **Use named parameters** for clarity:
   ```go
   db.NamedExec("INSERT INTO books (name, isbn) VALUES (:name, :isbn)", book)
   ```

2. **Handle NULL values** with pointers:
   ```go
   CompletedAt *time.Time `db:"completed_at"`
   ```

3. **Use transactions** for multi-step operations:
   ```go
   tx, _ := r.db.BeginTxx(ctx, nil)
   defer tx.Rollback()
   // operations...
   tx.Commit()
   ```

4. **Scan into domain entities** directly:
   ```go
   var books []book.Book
   db.Select(&books, query)
   ```

## Migration Support

SQL migrations in `migrations/postgres/`:
- Numbered sequence: `000001_create_books.up.sql`
- Up and down migrations
- Run via `make migrate-up`