package postgres

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// TestEntity is a test entity for generic repository tests
type TestEntity struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

// TestGetByID tests the GetByID generic function
func TestGetByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	testID := "test-id-123"
	expectedName := "Test Name"

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(testID, expectedName)
	mock.ExpectQuery("SELECT \\* FROM test_table WHERE id=\\$1").
		WithArgs(testID).
		WillReturnRows(rows)

	// Execute
	entity, err := GetByID[TestEntity](ctx, db, "test_table", testID)

	// Assert
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if entity.ID != testID {
		t.Errorf("GetByID() ID = %v, want %v", entity.ID, testID)
	}
	if entity.Name != expectedName {
		t.Errorf("GetByID() Name = %v, want %v", entity.Name, expectedName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestGetByIDWithColumns tests the GetByIDWithColumns generic function
func TestGetByIDWithColumns(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	testID := "test-id-456"
	expectedName := "Another Name"

	// Expect SELECT query with specific columns
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(testID, expectedName)
	mock.ExpectQuery("SELECT id, name FROM test_table WHERE id=\\$1").
		WithArgs(testID).
		WillReturnRows(rows)

	// Execute
	entity, err := GetByIDWithColumns[TestEntity](ctx, db, "test_table", "id, name", testID)

	// Assert
	if err != nil {
		t.Fatalf("GetByIDWithColumns() error = %v", err)
	}
	if entity.ID != testID {
		t.Errorf("GetByIDWithColumns() ID = %v, want %v", entity.ID, testID)
	}
	if entity.Name != expectedName {
		t.Errorf("GetByIDWithColumns() Name = %v, want %v", entity.Name, expectedName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestList tests the List generic function
func TestList(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()

	// Expect SELECT query
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("id-1", "Name 1").
		AddRow("id-2", "Name 2").
		AddRow("id-3", "Name 3")
	mock.ExpectQuery("SELECT \\* FROM test_table ORDER BY id").
		WillReturnRows(rows)

	// Execute
	entities, err := List[TestEntity](ctx, db, "test_table", "id")

	// Assert
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entities) != 3 {
		t.Errorf("List() returned %d entities, want 3", len(entities))
	}
	if entities[0].ID != "id-1" {
		t.Errorf("List() first entity ID = %v, want %v", entities[0].ID, "id-1")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestListWithColumns tests the ListWithColumns generic function
func TestListWithColumns(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()

	// Expect SELECT query with specific columns
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("id-1", "Name 1").
		AddRow("id-2", "Name 2")
	mock.ExpectQuery("SELECT id, name FROM test_table ORDER BY created_at DESC").
		WillReturnRows(rows)

	// Execute
	entities, err := ListWithColumns[TestEntity](ctx, db, "test_table", "id, name", "created_at DESC")

	// Assert
	if err != nil {
		t.Fatalf("ListWithColumns() error = %v", err)
	}
	if len(entities) != 2 {
		t.Errorf("ListWithColumns() returned %d entities, want 2", len(entities))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestListWithColumns_DefaultOrderBy tests that default ORDER BY is applied
func TestListWithColumns_DefaultOrderBy(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()

	// Expect SELECT query with default ORDER BY id
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("id-1", "Name 1")
	mock.ExpectQuery("SELECT id, name FROM test_table ORDER BY id").
		WillReturnRows(rows)

	// Execute with empty orderBy
	entities, err := ListWithColumns[TestEntity](ctx, db, "test_table", "id, name", "")

	// Assert
	if err != nil {
		t.Fatalf("ListWithColumns() error = %v", err)
	}
	if len(entities) != 1 {
		t.Errorf("ListWithColumns() returned %d entities, want 1", len(entities))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestDeleteByID tests the DeleteByID function
func TestDeleteByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	testID := "delete-id-789"

	// Expect DELETE query
	rows := sqlmock.NewRows([]string{"id"}).AddRow(testID)
	mock.ExpectQuery("DELETE FROM test_table WHERE id=\\$1 RETURNING id").
		WithArgs(testID).
		WillReturnRows(rows)

	// Execute
	err := DeleteByID(ctx, db, "test_table", testID)

	// Assert
	if err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

// TestExistsByID tests the ExistsByID function
func TestExistsByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	testID := "exists-id-999"

	tests := []struct {
		name     string
		exists   bool
		expected bool
	}{
		{"entity exists", true, true},
		{"entity does not exist", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Expect EXISTS query
			rows := sqlmock.NewRows([]string{"exists"}).AddRow(tt.exists)
			mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM test_table WHERE id=\\$1\\)").
				WithArgs(testID).
				WillReturnRows(rows)

			// Execute
			exists, err := ExistsByID(ctx, db, "test_table", testID)

			// Assert
			if err != nil {
				t.Fatalf("ExistsByID() error = %v", err)
			}
			if exists != tt.expected {
				t.Errorf("ExistsByID() = %v, want %v", exists, tt.expected)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unmet expectations: %v", err)
			}
		})
	}
}

// TestCountAll tests the CountAll function
func TestCountAll(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	ctx := context.Background()
	expectedCount := int64(42)

	// Expect COUNT query
	rows := sqlmock.NewRows([]string{"count"}).AddRow(expectedCount)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_table").
		WillReturnRows(rows)

	// Execute
	count, err := CountAll(ctx, db, "test_table")

	// Assert
	if err != nil {
		t.Fatalf("CountAll() error = %v", err)
	}
	if count != expectedCount {
		t.Errorf("CountAll() = %v, want %v", count, expectedCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
