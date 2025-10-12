package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"library-service/internal/infrastructure/store"
)

func TestHandleSQLError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantErr: nil,
		},
		{
			name:    "sql.ErrNoRows returns store.ErrorNotFound",
			err:     sql.ErrNoRows,
			wantErr: store.ErrorNotFound,
		},
		{
			name:    "wrapped sql.ErrNoRows returns store.ErrorNotFound",
			err:     errors.Join(errors.New("query failed"), sql.ErrNoRows),
			wantErr: store.ErrorNotFound,
		},
		{
			name:    "other error passed through",
			err:     errors.New("some database error"),
			wantErr: errors.New("some database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleSQLError(tt.err)

			// For nil case
			if tt.wantErr == nil {
				if got != nil {
					t.Errorf("HandleSQLError() = %v, want nil", got)
				}
				return
			}

			// For store.ErrorNotFound
			if tt.wantErr == store.ErrorNotFound {
				if got != store.ErrorNotFound {
					t.Errorf("HandleSQLError() = %v, want store.ErrorNotFound", got)
				}
				return
			}

			// For other errors, just check error message
			if got == nil {
				t.Errorf("HandleSQLError() = nil, want error")
				return
			}
			if got.Error() != tt.wantErr.Error() {
				t.Errorf("HandleSQLError() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func setupBaseRepoTest(t *testing.T) (*BaseRepository[TestEntity], sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	baseRepo := NewBaseRepository[TestEntity](sqlxDB, "test_table")

	return &baseRepo, mock
}

func TestNewBaseRepository(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	repo := NewBaseRepository[TestEntity](sqlxDB, "test_table")

	if repo.GetTableName() != "test_table" {
		t.Errorf("expected table name 'test_table', got '%s'", repo.GetTableName())
	}

	if repo.GetDB() != sqlxDB {
		t.Error("expected GetDB to return the same db instance")
	}
}

func TestBaseRepository_GenerateID(t *testing.T) {
	repo, _ := setupBaseRepoTest(t)

	id1 := repo.GenerateID()
	id2 := repo.GenerateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}

	if id1 == id2 {
		t.Error("expected different IDs, got duplicate")
	}

	// UUID format check (simple validation)
	if len(id1) != 36 {
		t.Errorf("expected UUID format (36 chars), got %d chars: %s", len(id1), id1)
	}
}

func TestBaseRepository_Get(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("test-id-123", "Test Entity")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table WHERE id=$1")).
			WithArgs("test-id-123").
			WillReturnRows(rows)

		entity, err := repo.Get(ctx, "test-id-123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if entity.ID != "test-id-123" {
			t.Errorf("expected ID 'test-id-123', got '%s'", entity.ID)
		}

		if entity.Name != "Test Entity" {
			t.Errorf("expected Name 'Test Entity', got '%s'", entity.Name)
		}
	})

	t.Run("entity not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table WHERE id=$1")).
			WithArgs("non-existent").
			WillReturnError(sqlmock.ErrCancelled)

		_, err := repo.Get(ctx, "non-existent")
		if err == nil {
			t.Error("expected error for non-existent entity")
		}
	})
}

func TestBaseRepository_List(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("list all entities", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("id-1", "Entity 1").
			AddRow("id-2", "Entity 2").
			AddRow("id-3", "Entity 3")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table ORDER BY id")).
			WillReturnRows(rows)

		entities, err := repo.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(entities) != 3 {
			t.Errorf("expected 3 entities, got %d", len(entities))
		}

		if entities[0].ID != "id-1" || entities[0].Name != "Entity 1" {
			t.Errorf("unexpected first entity: %+v", entities[0])
		}
	})

	t.Run("empty list", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table ORDER BY id")).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		entities, err := repo.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("expected empty list, got %d entities", len(entities))
		}
	})
}

func TestBaseRepository_ListWithOrder(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("id-3", "Entity 3").
		AddRow("id-2", "Entity 2").
		AddRow("id-1", "Entity 1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM test_table ORDER BY name DESC")).
		WillReturnRows(rows)

	entities, err := repo.ListWithOrder(ctx, "name DESC")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(entities) != 3 {
		t.Errorf("expected 3 entities, got %d", len(entities))
	}

	// Check descending order
	if entities[0].Name != "Entity 3" {
		t.Errorf("expected first entity name 'Entity 3', got '%s'", entities[0].Name)
	}
}

func TestBaseRepository_Delete(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow("test-id-123")

		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM test_table WHERE id=$1 RETURNING id")).
			WithArgs("test-id-123").
			WillReturnRows(rows)

		err := repo.Delete(ctx, "test-id-123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM test_table WHERE id=$1 RETURNING id")).
			WithArgs("bad-id").
			WillReturnError(sqlmock.ErrCancelled)

		err := repo.Delete(ctx, "bad-id")
		if err == nil {
			t.Error("expected error for failed delete")
		}
	})
}

func TestBaseRepository_Exists(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("entity exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM test_table WHERE id=$1)")).
			WithArgs("test-id-123").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "test-id-123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !exists {
			t.Error("expected entity to exist")
		}
	})

	t.Run("entity does not exist", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM test_table WHERE id=$1)")).
			WithArgs("non-existent").
			WillReturnRows(rows)

		exists, err := repo.Exists(ctx, "non-existent")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if exists {
			t.Error("expected entity to not exist")
		}
	})
}

func TestBaseRepository_Count(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(int64(42))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM test_table")).
		WillReturnRows(rows)

	count, err := repo.Count(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != int64(42) {
		t.Errorf("expected count 42, got %d", count)
	}
}

func TestBaseRepository_BatchGet(t *testing.T) {
	repo, _ := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("empty IDs list", func(t *testing.T) {
		entities, err := repo.BatchGet(ctx, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(entities) != 0 {
			t.Errorf("expected empty result for empty IDs, got %d entities", len(entities))
		}
	})

	// Note: Testing BatchGet with actual IDs requires a real PostgreSQL connection
	// due to the complexity of mocking PostgreSQL array types with sqlmock.
	// This method is better tested in integration tests.
}

func TestBaseRepository_Transaction(t *testing.T) {
	repo, mock := setupBaseRepoTest(t)
	defer repo.GetDB().Close()

	ctx := context.Background()

	t.Run("successful transaction", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE test_table").WillReturnResult(driver.ResultNoRows)
		mock.ExpectCommit()

		err := repo.Transaction(ctx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec("UPDATE test_table SET name = 'updated'")
			return err
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("transaction rollback on error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE test_table").WillReturnError(sqlmock.ErrCancelled)
		mock.ExpectRollback()

		err := repo.Transaction(ctx, func(tx *sqlx.Tx) error {
			_, err := tx.Exec("UPDATE test_table SET name = 'updated'")
			return err
		})

		if err == nil {
			t.Error("expected error for failed transaction")
		}
	})
}
