package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDatabase, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	databaseConnection, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDatabase}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	return databaseConnection, mock
}

func TestManager(t *testing.T) {
	databaseConnection, mock := newMockDatabase(t)
	manager := NewManager(databaseConnection)

	t.Run("commit", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := manager.Execute(context.Background(), func(transactionContext context.Context) error {
			resolved := DatabaseConnection(transactionContext, databaseConnection)
			if resolved == nil {
				t.Fatal("DatabaseConnection() = nil")
			}
			if resolved.Statement.ConnPool == databaseConnection.Statement.ConnPool {
				t.Fatal("DatabaseConnection() returned the fallback connection")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		expectedError := errors.New("operation failed")
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := manager.Execute(context.Background(), func(context.Context) error {
			return expectedError
		})
		if !errors.Is(err, expectedError) {
			t.Fatalf("Execute() error = %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseConnectionUsesFallback(t *testing.T) {
	databaseConnection, mock := newMockDatabase(t)
	resolved := DatabaseConnection(context.Background(), databaseConnection)
	if resolved == nil {
		t.Fatal("DatabaseConnection() = nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
