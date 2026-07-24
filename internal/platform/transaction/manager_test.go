package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
)

func newMockDatabase(t *testing.T) (*Manager, *databasequeries.Queries, sqlmock.Sqlmock) {
	t.Helper()

	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	return NewManager(databaseConnection), databasequeries.New(databaseConnection), mock
}

func TestManagerCommitsSuccessfulOperation(t *testing.T) {
	manager, fallbackQueries, mock := newMockDatabase(t)
	mock.ExpectBegin()
	mock.ExpectCommit()

	errorValue := manager.Execute(
		context.Background(),
		func(transactionContext context.Context) error {
			resolvedQueries := Queries(transactionContext, fallbackQueries)
			if resolvedQueries == fallbackQueries {
				t.Fatal("Queries() returned the fallback outside the SQL transaction")
			}
			return nil
		},
	)
	if errorValue != nil {
		t.Fatalf("Execute() error = %v", errorValue)
	}

	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestManagerRollsBackFailedOperation(t *testing.T) {
	manager, _, mock := newMockDatabase(t)
	expectedError := errors.New("operation failed")
	mock.ExpectBegin()
	mock.ExpectRollback()

	errorValue := manager.Execute(
		context.Background(),
		func(context.Context) error { return expectedError },
	)
	if !errors.Is(errorValue, expectedError) {
		t.Fatalf("Execute() error = %v", errorValue)
	}

	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestManagerReportsTransactionErrors(t *testing.T) {
	t.Run("begin", func(t *testing.T) {
		manager, _, mock := newMockDatabase(t)
		expectedError := errors.New("begin failed")
		mock.ExpectBegin().WillReturnError(expectedError)

		if errorValue := manager.Execute(context.Background(), func(context.Context) error {
			return nil
		}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("Execute() error = %v", errorValue)
		}
	})

	t.Run("rollback", func(t *testing.T) {
		manager, _, mock := newMockDatabase(t)
		operationError := errors.New("operation failed")
		rollbackError := errors.New("rollback failed")
		mock.ExpectBegin()
		mock.ExpectRollback().WillReturnError(rollbackError)

		if errorValue := manager.Execute(context.Background(), func(context.Context) error {
			return operationError
		}); !errors.Is(errorValue, rollbackError) {
			t.Fatalf("Execute() error = %v", errorValue)
		}
	})

	t.Run("commit", func(t *testing.T) {
		manager, _, mock := newMockDatabase(t)
		expectedError := errors.New("commit failed")
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(expectedError)

		if errorValue := manager.Execute(context.Background(), func(context.Context) error {
			return nil
		}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("Execute() error = %v", errorValue)
		}
	})
}

func TestQueriesUsesFallbackWithoutTransaction(t *testing.T) {
	_, fallbackQueries, _ := newMockDatabase(t)
	if resolvedQueries := Queries(context.Background(), fallbackQueries); resolvedQueries != fallbackQueries {
		t.Fatal("Queries() did not return the fallback")
	}
}
