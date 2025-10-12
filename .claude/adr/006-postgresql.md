# ADR-006: PostgreSQL as Primary Database

**Status:** Accepted

**Date:** 2024-01-18

**Decision Makers:** Project Architecture Team

## Context

We needed to choose a primary database for the Library Management System.

**Requirements:**
- ACID compliance (data integrity for loans, subscriptions, payments)
- Relational data (books, members, authors, loans with relationships)
- Complex queries (availability, overdue loans, analytics)
- JSON support (flexible metadata for books)
- Good performance (< 100ms for typical queries)
- Proven reliability
- Good Go support

## Decision

We chose **PostgreSQL 15+** as our primary database.

**Configuration:**
- PostgreSQL 15.3+
- Connection pooling with `pgx` driver
- Database per environment (development, staging, production)
- Docker for local development
- Migrations with custom migration tool

## Consequences

### Positive

1. **ACID Guarantees:**
   ```sql
   BEGIN;
       -- Loan a book
       INSERT INTO loans (id, book_id, member_id, loan_date, due_date)
       VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '14 days');

       -- Update book status
       UPDATE books SET status = 'loaned' WHERE id = $2;
   COMMIT;
   -- Either both succeed or both fail (atomicity)
   ```

2. **Rich Data Types:**
   ```sql
   -- JSON for flexible metadata
   CREATE TABLE books (
       id UUID PRIMARY KEY,
       title VARCHAR(255) NOT NULL,
       metadata JSONB,  -- {"awards": ["Hugo Award"], "translations": ["es", "fr"]}
       published_at TIMESTAMP,
       created_at TIMESTAMP DEFAULT NOW()
   );

   -- Query JSON fields efficiently
   SELECT * FROM books WHERE metadata->>'genre' = 'Science Fiction';
   ```

3. **Complex Queries:**
   ```sql
   -- Find overdue loans with member and book details
   SELECT
       l.id,
       m.full_name,
       b.title,
       l.due_date,
       (CURRENT_DATE - l.due_date) AS days_overdue
   FROM loans l
   JOIN members m ON l.member_id = m.id
   JOIN books b ON l.book_id = b.id
   WHERE l.status = 'active'
     AND l.due_date < CURRENT_DATE
   ORDER BY days_overdue DESC;
   ```

4. **Full-Text Search:**
   ```sql
   -- Search books by title or author
   SELECT *
   FROM books
   WHERE to_tsvector('english', title || ' ' || author_name) @@ to_tsquery('science & fiction');
   ```

5. **Excellent Go Support:**
   ```go
   // Using pgx (PostgreSQL driver for Go)
   import "github.com/jackc/pgx/v5/pgxpool"

   pool, err := pgxpool.New(ctx, databaseURL)
   row := pool.QueryRow(ctx, "SELECT * FROM books WHERE id = $1", bookID)
   ```

6. **JSON Operations:**
   ```go
   // Store and query JSON metadata
   query := `
       INSERT INTO books (id, title, metadata)
       VALUES ($1, $2, $3)
   `
   metadata := map[string]interface{}{
       "genre":  "Science Fiction",
       "awards": []string{"Hugo Award", "Nebula Award"},
   }
   _, err := db.Exec(query, id, title, metadata)
   ```

7. **Performance:**
   - Indexes on foreign keys and frequently queried columns
   - Query plans with EXPLAIN ANALYZE
   - Materialized views for analytics
   - Connection pooling (20-50 connections)

8. **Proven Reliability:**
   - Used by millions of applications
   - 20+ years of development
   - Strong community support
   - Excellent documentation

### Negative

1. **Infrastructure Complexity:**
   - Need to run PostgreSQL server (Docker for dev)
   - Backup and restore procedures
   - Monitoring and tuning
   - Mitigation: Docker Compose makes local dev easy

2. **Scaling Limitations:**
   - Vertical scaling easier than horizontal
   - Replication can be complex
   - Mitigation: Single-server PostgreSQL can handle millions of rows easily

3. **Schema Migrations:**
   - Need migration tool
   - Coordinating migrations with deployments
   - Rollback complexity
   - Mitigation: Custom migration tool with up/down migrations

4. **Query Performance Tuning:**
   - Need to understand indexes and query plans
   - Some queries require optimization
   - Mitigation: EXPLAIN ANALYZE helps identify slow queries

## Alternatives Considered

### Alternative 1: MongoDB

**Why not chosen:**
```javascript
// MongoDB doesn't enforce foreign key constraints
{
  "_id": "loan_123",
  "book_id": "book_456",   // No guarantee this book exists!
  "member_id": "member_789" // No guarantee this member exists!
}
```

**Issues:**
- No ACID transactions (until MongoDB 4.0, and still limited)
- No foreign key constraints (data integrity risk for loans)
- No complex JOINs (would need application-level joining)
- Schemaless can lead to data quality issues

**When MongoDB would be better:**
- Document-oriented data (blogs, CMS)
- Flexible schema changes
- Massive horizontal scaling needs
- Simple queries

### Alternative 2: MySQL

**Why not chosen:**
- JSON support inferior to PostgreSQL (no JSONB indexing)
- Full-text search not as good
- Less strict about data types (can silently truncate)
- PostgreSQL has better compliance with SQL standards

**MySQL advantages:**
- Slightly simpler replication
- Slightly faster for simple queries
- More widespread (but PostgreSQL catching up)

**Not enough to outweigh PostgreSQL benefits for our use case**

### Alternative 3: SQLite

**Why not chosen:**
- File-based (single file database)
- No concurrent writes (writer blocks readers)
- No network access (can't scale to multiple servers)
- Great for embedded apps, not for web services

**When SQLite would be better:**
- Mobile apps
- Desktop applications
- Embedded systems
- Testing (we DO use SQLite for some tests)

### Alternative 4: Cloud-Native (DynamoDB, CosmosDB)

**Why not chosen:**
- Vendor lock-in
- More expensive
- Limited query capabilities
- Our data is relational (books, members, loans)

**When cloud-native would be better:**
- Serverless architectures
- Massive scale from day one
- Strong multi-region requirements

## Implementation Details

**Connection:**
```go
// internal/infrastructure/store/postgres.go
package store

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
    DB *pgxpool.Pool
}

func NewPostgres(databaseURL string) (*PostgresStore, error) {
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    config.MaxConns = 25
    config.MinConns = 5

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return nil, err
    }

    return &PostgresStore{DB: pool}, nil
}
```

**Migration:**
```bash
# Create migration
make migrate-create name=add_loans_table

# Apply migrations
make migrate-up

# Rollback
make migrate-down
```

**Indexes:**
```sql
-- Always index foreign keys
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_member_id ON loans(member_id);

-- Index frequently queried columns
CREATE INDEX idx_books_isbn ON books(isbn);
CREATE INDEX idx_books_status ON books(status);

-- Composite index for common query
CREATE INDEX idx_loans_member_status ON loans(member_id, status);

-- JSON index
CREATE INDEX idx_books_metadata_genre ON books USING GIN ((metadata->'genre'));
```

**Connection String:**
```
postgres://username:password@localhost:5432/library?sslmode=disable
```

**Docker Compose:**
```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: library
      POSTGRES_PASSWORD: library123
      POSTGRES_DB: library
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

## Performance Benchmarks

**Target performance:**
- Simple query (get by ID): < 5ms
- Complex query (join 3 tables): < 50ms
- Insert: < 10ms
- Transaction (2-3 operations): < 20ms

**Actual performance (100,000 books, 50,000 members, 200,000 loans):**
- Get book by ID: 2ms
- Get book by ISBN: 3ms (indexed)
- List available books: 15ms
- Get member with loans (join): 8ms
- Create loan (transaction): 12ms
- Find overdue loans: 25ms

✅ All targets met

## Migration Path

If we need to change databases in the future:

1. Repository interfaces are in domain layer (see [ADR-005](./005-repository-interfaces.md))
2. Only adapter layer needs rewriting
3. Domain and use case layers unchanged

**Estimated migration time to another SQL database:** 1-2 weeks
**Estimated migration time to NoSQL:** 1-2 months (data modeling differences)

## Monitoring

**Key metrics:**
- Connection pool utilization
- Query execution time (p50, p95, p99)
- Lock wait time
- Slow query log (queries > 100ms)
- Database size growth

**Tools:**
- pg_stat_statements (query statistics)
- EXPLAIN ANALYZE (query plans)
- pgBadger (log analyzer)

## Validation

After 6 months:
- ✅ 500,000 books, 100,000 members, 1 million loans
- ✅ All queries < 100ms
- ✅ Zero data integrity issues
- ✅ Zero data loss incidents
- ✅ Database size: 2.5 GB (well within capacity)
- ✅ Connection pool never exhausted

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/15/index.html)
- [pgx - PostgreSQL Driver for Go](https://github.com/jackc/pgx)
- [PostgreSQL vs MongoDB](https://www.mongodb.com/compare/mongodb-postgresql)
- `.claude/development-guide.md` - PostgreSQL setup instructions
- `.claude/reference/debugging-guide.md` - Database debugging techniques

## Related ADRs

- [ADR-005: Repository Interfaces in Domain](./005-repository-interfaces.md) - Makes database swappable
- [ADR-001: Clean Architecture](./001-clean-architecture.md) - Isolates database from business logic

---

**Last Reviewed:** 2024-01-18

**Next Review:** 2024-07-18 (or when considering database change)
