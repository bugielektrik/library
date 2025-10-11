# Database Migrations

SQL migration files for database schema management.

## Structure

```
migrations/
├── postgres/    # PostgreSQL migrations
│   ├── 000001_create_members_table.up.sql
│   ├── 000001_create_members_table.down.sql
│   ├── 000002_create_books_table.up.sql
│   ├── 000002_create_books_table.down.sql
│   └── ...
└── README.md
```

## Naming Convention

```
{sequence}_{description}.{direction}.sql
```

- **sequence**: 6-digit number (000001, 000002, ...)
- **description**: Snake_case description
- **direction**: `up` (apply) or `down` (rollback)

Examples:
- `000001_create_members_table.up.sql`
- `000001_create_members_table.down.sql`

## Creating Migrations

```bash
# Create new migration pair
make migrate-create name=add_payment_status

# This creates:
# - migrations/postgres/000XXX_add_payment_status.up.sql
# - migrations/postgres/000XXX_add_payment_status.down.sql
```

## Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# With custom database URL
POSTGRES_DSN="postgres://user:pass@localhost/db" make migrate-up
```

## Best Practices

### 1. Always Create Both Files
Every `.up.sql` must have a corresponding `.down.sql`:
```sql
-- 000001_create_books.up.sql
CREATE TABLE books (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 000001_create_books.down.sql
DROP TABLE IF EXISTS books;
```

### 2. Make Migrations Idempotent
Use `IF NOT EXISTS` and `IF EXISTS`:
```sql
CREATE TABLE IF NOT EXISTS books (...);
CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);
```

### 3. Never Modify Applied Migrations
Once applied to production, migrations are immutable. Create new migrations for changes.

### 4. Keep Migrations Small
One logical change per migration:
- ❌ `000001_initial_schema.sql` (too large)
- ✅ `000001_create_members.sql`
- ✅ `000002_create_books.sql`

### 5. Test Rollbacks
Always test that down migrations work:
```bash
make migrate-up
make migrate-down
make migrate-up
```

## Common Patterns

### Adding a Column
```sql
-- up
ALTER TABLE payments
ADD COLUMN refund_amount INTEGER;

-- down
ALTER TABLE payments
DROP COLUMN refund_amount;
```

### Creating an Index
```sql
-- up
CREATE INDEX CONCURRENTLY idx_payments_member_id
ON payments(member_id);

-- down
DROP INDEX IF EXISTS idx_payments_member_id;
```

### Adding a Constraint
```sql
-- up
ALTER TABLE books
ADD CONSTRAINT uk_books_isbn UNIQUE(isbn);

-- down
ALTER TABLE books
DROP CONSTRAINT IF EXISTS uk_books_isbn;
```

## Migration Tool

Using [golang-migrate](https://github.com/golang-migrate/migrate):
- Tracks applied migrations in `schema_migrations` table
- Ensures migrations run once and in order
- Supports rollback to specific version

## Troubleshooting

### Migration Failed
```bash
# Check current version
psql -c "SELECT * FROM schema_migrations"

# Force version (use with caution)
migrate -path migrations/postgres -database $POSTGRES_DSN force VERSION
```

### Dirty Database
If migration fails midway, database may be marked dirty:
```bash
# Fix by forcing version
migrate force VERSION

# Then continue
make migrate-up
```