# Integration Testing Guide

This guide shows how to write integration tests that use real PostgreSQL database.

**Time Estimate:** 30 minutes to set up, 15 minutes per test suite

---

## Overview

**Integration tests:**
- Use real PostgreSQL database (via Docker)
- Test full request→database→response flow
- Verify actual SQL queries work
- Catch integration bugs before production

**Build tag:** `//go:build integration`

---

## Step 1: Database Setup (One-time, 10 minutes)

### 1.1 Test Database Configuration

**File:** `test/testdb/setup.go`

```go
package testdb

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"library-service/internal/infrastructure/store"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Test database connection string
	dsn := "postgres://library:library123@localhost:5432/library_test?sslmode=disable"

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// CleanupTestDB closes database connection
func CleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// TruncateTables clears all data from tables (for test isolation)
func TruncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// WithTransaction runs test within a transaction (auto-rollback)
func WithTransaction(t *testing.T, db *sqlx.DB, fn func(tx *sqlx.Tx)) {
	t.Helper()

	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			t.Logf("Failed to rollback transaction: %v", err)
		}
	}()

	fn(tx)
}
```

---

## Step 2: Test Fixtures (One-time, 15 minutes)

**File:** `test/fixtures/books.go`

```go
package fixtures

import (
	"time"

	"library-service/internal/domain/book"
)

// CreateTestBook returns a valid test book entity
func CreateTestBook(overrides ...func(*book.Book)) book.Book {
	name := "Test Book"
	genre := "Fiction"
	isbn := "978-0-306-40615-7"

	b := book.Book{
		ID:        "test-book-1",
		Name:      &name,
		Genre:     &genre,
		ISBN:      &isbn,
		Authors:   []string{"author-1"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(&b)
	}

	return b
}

// CreateTestBooks returns multiple test books
func CreateTestBooks(count int) []book.Book {
	books := make([]book.Book, count)
	for i := 0; i < count; i++ {
		books[i] = CreateTestBook(func(b *book.Book) {
			id := fmt.Sprintf("test-book-%d", i+1)
			name := fmt.Sprintf("Test Book %d", i+1)
			b.ID = id
			b.Name = &name
		})
	}
	return books
}
```

**File:** `test/fixtures/members.go`

```go
package fixtures

import (
	"time"

	"library-service/internal/domain/member"
)

// CreateTestMember returns a valid test member entity
func CreateTestMember(overrides ...func(*member.Member)) member.Member {
	fullName := "Test User"

	m := member.Member{
		ID:           "test-member-1",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hash...", // bcrypt hash of "password123"
		FullName:     &fullName,
		Role:         member.RoleUser,
		Books:        []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(&m)
	}

	return m
}
```

---

## Step 3: Writing Integration Tests

### 3.1 Repository Integration Test

**File:** `internal/adapters/repository/postgres/book_integration_test.go`

```go
//go:build integration

package postgres

import (
	"context"
	"testing"

	"library-service/internal/domain/book"
	"library-service/test/fixtures"
	"library-service/test/testdb"
)

func TestBookRepository_Create_Integration(t *testing.T) {
	// Setup
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Clean tables before test
	testdb.TruncateTables(t, db, "books", "authors")

	// Test data
	testBook := fixtures.CreateTestBook()

	// Execute
	bookID, err := repo.Add(ctx, testBook)

	// Assert
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if bookID == "" {
		t.Error("Create() returned empty ID")
	}

	// Verify book was actually saved
	savedBook, err := repo.Get(ctx, bookID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if savedBook.ID != bookID {
		t.Errorf("Got book ID = %v, want %v", savedBook.ID, bookID)
	}

	if *savedBook.Name != *testBook.Name {
		t.Errorf("Got book name = %v, want %v", *savedBook.Name, *testBook.Name)
	}
}

func TestBookRepository_List_Integration(t *testing.T) {
	// Setup
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	repo := NewBookRepository(db)
	ctx := context.Background()

	// Clean tables
	testdb.TruncateTables(t, db, "books")

	// Create test books
	testBooks := fixtures.CreateTestBooks(3)
	for _, b := range testBooks {
		if _, err := repo.Add(ctx, b); err != nil {
			t.Fatalf("Failed to create test book: %v", err)
		}
	}

	// Execute
	books, err := repo.List(ctx)

	// Assert
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(books) != 3 {
		t.Errorf("List() returned %d books, want 3", len(books))
	}
}
```

---

### 3.2 Use Case Integration Test

**File:** `internal/usecase/bookops/create_book_integration_test.go`

```go
//go:build integration

package bookops

import (
	"context"
	"testing"

	"library-service/internal/adapters/cache/memory"
	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/domain/book"
	"library-service/test/fixtures"
	"library-service/test/testdb"
)

func TestCreateBookUseCase_Integration(t *testing.T) {
	// Setup database
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	// Clean tables
	testdb.TruncateTables(t, db, "books", "authors")

	// Setup real dependencies
	bookRepo := postgres.NewBookRepository(db)
	bookCache := memory.NewBookCache()
	bookService := book.NewService()

	// Create use case with real dependencies
	uc := NewCreateBookUseCase(bookRepo, bookCache, bookService)

	ctx := context.Background()

	// Test data
	req := CreateBookRequest{
		Name:    "Integration Test Book",
		Genre:   "Technical",
		ISBN:    "978-0-306-40615-7",
		Authors: []string{},
	}

	// Execute
	result, err := uc.Execute(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ID == "" {
		t.Error("Execute() returned empty book ID")
	}

	// Verify book exists in database
	savedBook, err := bookRepo.Get(ctx, result.ID)
	if err != nil {
		t.Fatalf("Failed to get created book: %v", err)
	}

	if *savedBook.Name != req.Name {
		t.Errorf("Got name = %v, want %v", *savedBook.Name, req.Name)
	}

	// Verify book exists in cache
	cachedBook, err := bookCache.Get(ctx, result.ID)
	if err != nil {
		t.Error("Book should be in cache")
	}

	if cachedBook.ID != result.ID {
		t.Error("Cached book ID doesn't match")
	}
}
```

---

### 3.3 HTTP Handler Integration Test

**File:** `internal/adapters/http/handlers/book_integration_test.go`

```go
//go:build integration

package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/usecase"
	"library-service/test/testdb"
)

func TestBookHandler_Create_Integration(t *testing.T) {
	// Setup database
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	// Clean tables
	testdb.TruncateTables(t, db, "books")

	// Setup real dependencies
	bookRepo := postgres.NewBookRepository(db)
	// ... setup other dependencies ...

	// Create handler with real use cases
	handler := NewBookHandler(useCases, validator)

	// Test request
	reqBody := dto.CreateBookRequest{
		Name:    "Integration Test Book",
		Genre:   "Fiction",
		ISBN:    "978-0-306-40615-7",
		Authors: []string{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Execute
	rr := httptest.NewRecorder()
	handler.create(rr, req)

	// Assert
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response dto.BookResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID == "" {
		t.Error("Response missing book ID")
	}

	if response.Name != reqBody.Name {
		t.Errorf("Got name = %v, want %v", response.Name, reqBody.Name)
	}
}
```

---

## Step 4: Running Integration Tests

### 4.1 Setup Test Database

```bash
# Start PostgreSQL (if not running)
make up

# Create test database
psql -h localhost -U library -d library << EOF
CREATE DATABASE library_test;
EOF

# Run migrations on test database
POSTGRES_DSN="postgres://library:library123@localhost:5432/library_test?sslmode=disable" \
  go run cmd/migrate/main.go up
```

### 4.2 Run Integration Tests

```bash
# Run all integration tests
go test ./... -tags=integration -v

# Run specific package integration tests
go test ./internal/adapters/repository/postgres/... -tags=integration -v

# Run with coverage
go test ./... -tags=integration -coverprofile=coverage-integration.out

# Run integration + unit tests
go test ./... -tags=integration -v
```

### 4.3 Makefile Integration

**Add to Makefile:**

```makefile
# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test ./... -tags=integration -v -count=1

# Setup test database
.PHONY: test-db-setup
test-db-setup:
	@echo "Setting up test database..."
	psql -h localhost -U library -d library -c "CREATE DATABASE IF NOT EXISTS library_test;"
	POSTGRES_DSN="postgres://library:library123@localhost:5432/library_test?sslmode=disable" \
		go run cmd/migrate/main.go up

# Clean test database
.PHONY: test-db-clean
test-db-clean:
	@echo "Cleaning test database..."
	psql -h localhost -U library -d library -c "DROP DATABASE IF EXISTS library_test;"
```

---

## Step 5: Best Practices

### 5.1 Test Isolation

**Use transactions for automatic rollback:**

```go
func TestWithTransaction(t *testing.T) {
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	testdb.WithTransaction(t, db, func(tx *sqlx.Tx) {
		// All database operations within this function
		// will be rolled back automatically

		repo := NewBookRepository(db)
		// ... run your test ...

		// No need to clean up - transaction rolls back
	})
}
```

### 5.2 Parallel Tests

```go
func TestBookRepository_Create(t *testing.T) {
	t.Parallel() // Run in parallel with other tests

	// Each test gets its own database connection
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	// ...
}
```

### 5.3 Table-Driven Integration Tests

```go
func TestBookRepository_GetByISBN_Integration(t *testing.T) {
	db := testdb.SetupTestDB(t)
	defer testdb.CleanupTestDB(t, db)

	repo := NewBookRepository(db)
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func() // Setup test data
		isbn       string
		wantErr    bool
		wantResult bool
	}{
		{
			name: "book exists",
			setup: func() {
				book := fixtures.CreateTestBook()
				repo.Add(ctx, book)
			},
			isbn:       "978-0-306-40615-7",
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "book not found",
			setup:      func() {},
			isbn:       "999-9-999-99999-9",
			wantErr:    true,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean before each subtest
			testdb.TruncateTables(t, db, "books")

			// Setup
			tt.setup()

			// Execute
			result, err := repo.GetByISBN(ctx, tt.isbn)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByISBN() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantResult && result == nil {
				t.Error("Expected result, got nil")
			}
		})
	}
}
```

---

## Step 6: CI/CD Integration

**File:** `.github/workflows/integration-tests.yml`

```yaml
name: Integration Tests

on:
  pull_request:
    branches: [main, develop]

jobs:
  integration:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: library
          POSTGRES_PASSWORD: library123
          POSTGRES_DB: library_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run migrations
        run: |
          POSTGRES_DSN="postgres://library:library123@localhost:5432/library_test?sslmode=disable" \
            go run cmd/migrate/main.go up

      - name: Run integration tests
        run: go test ./... -tags=integration -v -race -coverprofile=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

---

## Summary

**Integration Testing Checklist:**

- ✅ **Setup:** Test database configuration
- ✅ **Fixtures:** Reusable test data
- ✅ **Tests:** Repository, use case, handler tests
- ✅ **Isolation:** Clean tables or use transactions
- ✅ **Run:** Local and CI/CD execution
- ✅ **Coverage:** Track integration test coverage

**When to Write Integration Tests:**

✅ **Write for:**
- Complex SQL queries
- Multi-table operations
- Transaction handling
- Repository implementations

❌ **Don't write for:**
- Domain logic (use unit tests)
- Simple CRUD operations (covered by other tests)
- Pure functions (use unit tests)

**Time Investment:**
- Setup: 30 minutes (one-time)
- Per test suite: 15-30 minutes
- Maintenance: Minimal (tests catch real bugs!)

**Integration tests give you confidence that your code works with real databases in production!** 🚀
