package inventoryrepository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
)

var errStockDatabase = errors.New("stock database failed")

func newStockMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDatabase, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	database, errorValue := gorm.Open(postgres.New(postgres.Config{Conn: sqlDatabase}), &gorm.Config{DisableAutomaticPing: true})
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return database, mock
}

func TestStockModelTableName(t *testing.T) {
	if (StockModel{}).TableName() != "stocks" {
		t.Fatalf("TableName() = %q", (StockModel{}).TableName())
	}
}

func TestStockRepositorySave(t *testing.T) {
	stock, _ := domain.NewStock("product-1", 5, time.Now())
	for _, testCase := range []struct {
		name      string
		dbError   error
		wantError error
	}{{"success", nil, nil}, {"database error", errStockDatabase, errStockDatabase}} {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newStockMockDatabase(t)
			mock.ExpectBegin()
			expectation := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stocks"`))
			if testCase.dbError != nil {
				expectation.WillReturnError(testCase.dbError)
				mock.ExpectRollback()
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}
			if errorValue := NewStockRepository(database).Save(context.Background(), stock); !errors.Is(errorValue, testCase.wantError) {
				t.Fatalf("Save() error = %v", errorValue)
			}
		})
	}
}

func TestStockRepositoryFind(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	query := `SELECT .* FROM "stocks" WHERE product_id =`

	t.Run("not found", func(t *testing.T) {
		database, mock := newStockMockDatabase(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}))
		stock, errorValue := NewStockRepository(database).FindByProductID(context.Background(), "missing")
		if stock != nil || !errors.Is(errorValue, domain.ErrStockNotFound) {
			t.Fatalf("FindByProductID() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("database error with lock", func(t *testing.T) {
		database, mock := newStockMockDatabase(t)
		mock.ExpectQuery(query + `.*FOR UPDATE`).WillReturnError(errStockDatabase)
		stock, errorValue := NewStockRepository(database).FindByProductIDForUpdate(context.Background(), "product-1")
		if stock != nil || !errors.Is(errorValue, errStockDatabase) {
			t.Fatalf("FindByProductIDForUpdate() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("invalid stored stock", func(t *testing.T) {
		database, mock := newStockMockDatabase(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}).AddRow("product-1", -1, now))
		stock, errorValue := NewStockRepository(database).FindByProductID(context.Background(), "product-1")
		if stock != nil || !errors.Is(errorValue, domain.ErrInvalidStockQuantity) {
			t.Fatalf("FindByProductID() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("success with lock", func(t *testing.T) {
		database, mock := newStockMockDatabase(t)
		mock.ExpectQuery(query + `.*FOR UPDATE`).WillReturnRows(sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}).AddRow("product-1", 5, now))
		stock, errorValue := NewStockRepository(database).FindByProductIDForUpdate(context.Background(), "product-1")
		if errorValue != nil || stock.ProductID != "product-1" || stock.Quantity != 5 || !stock.UpdatedAt.Equal(now) {
			t.Fatalf("FindByProductIDForUpdate() = %#v, %v", stock, errorValue)
		}
	})
}
