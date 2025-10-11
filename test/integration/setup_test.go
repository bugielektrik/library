//go:build integration
// +build integration

package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection and returns a cleanup function
func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	// Get database DSN from environment or use default
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://library:library123@localhost:5432/library_test?sslmode=disable"
	}

	// Connect to postgres database to create test database
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "failed to connect to test database")

	// Ping to verify connection
	err = db.Ping()
	require.NoError(t, err, "failed to ping test database")

	// Wrap with sqlx
	sqlxDB := sqlx.NewDb(db, "postgres")

	// Clean up existing test data
	cleanupTestData(t, sqlxDB)

	// Return cleanup function
	cleanup := func() {
		cleanupTestData(t, sqlxDB)
		sqlxDB.Close()
	}

	return sqlxDB, cleanup
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Delete in reverse order of foreign key constraints
	tables := []string{
		"callback_retries",
		"saved_cards",
		"payments",
		"reservations",
		"members",
		"books",
		"authors",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		_, err := db.Exec(query)
		if err != nil {
			// Log warning but don't fail - table might not exist yet
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// setupTestEnv sets up environment variables for testing
func setupTestEnv(t *testing.T) {
	t.Helper()

	// Set test environment variables
	os.Setenv("APP_MODE", "test")
	os.Setenv("JWT_SECRET", "test-secret-key-for-integration-tests")
	os.Setenv("JWT_EXPIRY", "1h")

	// Payment gateway test config
	os.Setenv("EPAYMENT_BASE_URL", "https://test-api.edomain.kz")
	os.Setenv("EPAYMENT_CLIENT_ID", "test-client-id")
	os.Setenv("EPAYMENT_CLIENT_SECRET", "test-client-secret")
	os.Setenv("EPAYMENT_TERMINAL", "test-terminal")
}
