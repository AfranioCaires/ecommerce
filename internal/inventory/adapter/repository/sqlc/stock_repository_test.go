package inventoryrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

func newStockRepository(t *testing.T) (*StockRepository, sqlmock.Sqlmock) {
	t.Helper()
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	return NewStockRepository(databasequeries.New(databaseConnection)), mock
}

func TestStockRepositoryUpsertsQuantity(t *testing.T) {
	repository, mock := newStockRepository(t)
	updatedAt := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	stock, _ := domain.NewStock("product-1", 8, updatedAt)

	mock.ExpectExec("INSERT INTO stocks").
		WithArgs("product-1", int32(8), updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if errorValue := repository.Save(context.Background(), stock); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestStockRepositoryLocksStockForUpdate(t *testing.T) {
	repository, mock := newStockRepository(t)
	updatedAt := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("FOR UPDATE").
		WithArgs("product-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"product_id",
			"quantity",
			"updated_at",
		}).AddRow("product-1", int32(8), updatedAt))

	stock, errorValue := repository.FindByProductIDForUpdate(
		context.Background(),
		"product-1",
	)
	if errorValue != nil || stock.Quantity != 8 {
		t.Fatalf("FindByProductIDForUpdate() = %#v, %v", stock, errorValue)
	}

	mock.ExpectQuery("FROM stocks").WithArgs("missing").WillReturnError(sql.ErrNoRows)
	if _, errorValue := repository.FindByProductID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrStockNotFound) {
		t.Fatalf("FindByProductID() error = %v", errorValue)
	}
}

func TestStockRepositoryUpdatesReservedQuantityInTransaction(t *testing.T) {
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	repository := NewStockRepository(databasequeries.New(databaseConnection))
	manager := transaction.NewManager(databaseConnection)
	updatedAt := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("FOR UPDATE").
		WithArgs("product-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"product_id",
			"quantity",
			"updated_at",
		}).AddRow("product-1", int32(8), updatedAt))
	mock.ExpectExec("INSERT INTO stocks").
		WithArgs("product-1", int32(5), updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	errorValue = manager.Execute(context.Background(), func(transactionContext context.Context) error {
		stock, findError := repository.FindByProductIDForUpdate(
			transactionContext,
			"product-1",
		)
		if findError != nil {
			return findError
		}
		stock.Quantity = 5
		return repository.Save(transactionContext, stock)
	})
	if errorValue != nil {
		t.Fatalf("Execute() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}
