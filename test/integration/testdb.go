//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDB wraps a database connection for integration tests
type TestDB struct {
	*sqlx.DB
	t *testing.T
}

// Setup creates a test database connection
func Setup(t *testing.T) *TestDB {
	t.Helper()

	dsn := getTestDSN()
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Set connection pool settings for tests
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return &TestDB{DB: db, t: t}
}

// Cleanup closes the database connection and cleans up test data
func (db *TestDB) Cleanup() {
	if db.DB != nil {
		db.DB.Close()
	}
}

// Truncate removes all data from specified tables
func (db *TestDB) Truncate(tables ...string) {
	db.t.Helper()

	ctx := context.Background()
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.ExecContext(ctx, query); err != nil {
			db.t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

// TruncateAll removes all data from all test tables
func (db *TestDB) TruncateAll() {
	db.t.Helper()

	db.Truncate(
		"reservations",
		"saved_cards",
		"receipts",
		"payments",
		"callback_retries",
		"books",
		"authors",
		"members",
	)
}

// BeginTx starts a transaction for testing
func (db *TestDB) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return db.DB.BeginTxx(ctx, nil)
}

// Exec executes a query and fails the test if there's an error
func (db *TestDB) MustExec(query string, args ...interface{}) {
	db.t.Helper()

	if _, err := db.DB.Exec(query, args...); err != nil {
		db.t.Fatalf("failed to execute query: %v\nQuery: %s", err, query)
	}
}

// Get retrieves a single row and fails the test if there's an error
func (db *TestDB) MustGet(dest interface{}, query string, args ...interface{}) {
	db.t.Helper()

	if err := db.DB.Get(dest, query, args...); err != nil {
		db.t.Fatalf("failed to get row: %v\nQuery: %s", err, query)
	}
}

// Select retrieves multiple rows and fails the test if there's an error
func (db *TestDB) MustSelect(dest interface{}, query string, args ...interface{}) {
	db.t.Helper()

	if err := db.DB.Select(dest, query, args...); err != nil {
		db.t.Fatalf("failed to select rows: %v\nQuery: %s", err, query)
	}
}

// getTestDSN returns the test database connection string
func getTestDSN() string {
	// Default test database configuration
	// Override with POSTGRES_DSN_TEST environment variable for custom setup
	dsn := "postgres://library:library123@localhost:5432/library?sslmode=disable"

	// In CI/CD, you might want to use a different database
	// if envDSN := os.Getenv("POSTGRES_DSN_TEST"); envDSN != "" {
	// 	dsn = envDSN
	// }

	return dsn
}

// AssertRowCount checks that a table has the expected number of rows
func (db *TestDB) AssertRowCount(table string, expected int) {
	db.t.Helper()

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := db.DB.Get(&count, query); err != nil {
		db.t.Fatalf("failed to count rows in %s: %v", table, err)
	}

	if count != expected {
		db.t.Errorf("expected %d rows in %s, got %d", expected, table, count)
	}
}

// AssertExists checks that a row exists with the given ID
func (db *TestDB) AssertExists(table, id string) {
	db.t.Helper()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1)", table)
	if err := db.DB.Get(&exists, query, id); err != nil {
		db.t.Fatalf("failed to check existence in %s: %v", table, err)
	}

	if !exists {
		db.t.Errorf("expected row with id %s to exist in %s", id, table)
	}
}

// AssertNotExists checks that a row does not exist with the given ID
func (db *TestDB) AssertNotExists(table, id string) {
	db.t.Helper()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id=$1)", table)
	if err := db.DB.Get(&exists, query, id); err != nil {
		db.t.Fatalf("failed to check existence in %s: %v", table, err)
	}

	if exists {
		db.t.Errorf("expected row with id %s to not exist in %s", id, table)
	}
}
