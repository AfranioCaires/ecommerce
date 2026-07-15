package orderrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/order/usecase"
)

func newMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDatabase, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	databaseConnection, errorValue := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDatabase}),
		&gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true},
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return databaseConnection, mock
}

func sampleOrder(t *testing.T) *domain.Order {
	t.Helper()
	order, errorValue := domain.NewOrder(
		"order-1",
		"user-1",
		[]domain.OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1500, Quantity: 2}},
		time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return order
}

func orderRows(status string) *sqlmock.Rows {
	createdAt := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	return sqlmock.NewRows([]string{"id", "user_id", "total_amount_cents", "status", "created_at", "updated_at"}).
		AddRow("order-1", "user-1", 3000, status, createdAt, createdAt)
}

func itemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "order_id", "product_id", "product_name", "unit_price_cents", "quantity"}).
		AddRow(1, "order-1", "product-1", "Keyboard", 1500, 2)
}

func TestOrderModelTableNames(t *testing.T) {
	if tableName := (OrderModel{}).TableName(); tableName != "orders" {
		t.Fatalf("OrderModel.TableName() = %q", tableName)
	}
	if tableName := (OrderItemModel{}).TableName(); tableName != "order_items" {
		t.Fatalf("OrderItemModel.TableName() = %q", tableName)
	}
}

func TestOrderRepositorySaveAndUpdateStatus(t *testing.T) {
	databaseConnection, mock := newMockDatabase(t)
	repository := NewOrderRepository(databaseConnection)
	order := sampleOrder(t)

	mock.ExpectExec(`INSERT INTO "orders"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO "order_items"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	if errorValue := repository.Save(context.Background(), order); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}

	order.MarkAsPaid(order.UpdatedAt.Add(time.Hour))
	mock.ExpectExec(`UPDATE "orders" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if errorValue := repository.UpdateStatus(context.Background(), order); errorValue != nil {
		t.Fatalf("UpdateStatus() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestOrderRepositoryFindByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE id = \$1`).WithArgs("order-1", 1).WillReturnRows(orderRows(string(domain.OrderStatusPending)))
		mock.ExpectQuery(`SELECT \* FROM "order_items" WHERE "order_items"\."order_id" = \$1`).WithArgs("order-1").WillReturnRows(itemRows())

		order, errorValue := repository.FindByID(context.Background(), "order-1")
		if errorValue != nil || order.ID != "order-1" || len(order.Items) != 1 {
			t.Fatalf("FindByID() = %#v, %v", order, errorValue)
		}
		if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
			t.Fatal(errorValue)
		}
	})

	t.Run("not found", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE id = \$1`).WillReturnError(gorm.ErrRecordNotFound)
		if _, errorValue := repository.FindByID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrOrderNotFound) {
			t.Fatalf("FindByID() error = %v", errorValue)
		}
	})

	t.Run("database error", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		expectedError := errors.New("query failed")
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE id = \$1`).WillReturnError(expectedError)
		if _, errorValue := repository.FindByID(context.Background(), "order-1"); !errors.Is(errorValue, expectedError) {
			t.Fatalf("FindByID() error = %v", errorValue)
		}
	})
}

func TestOrderRepositoryFindMany(t *testing.T) {
	pageRequest := usecase.OrderPageRequest{Limit: 20, Offset: 0}

	t.Run("by user ID", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
			WithArgs("user-1", pageRequest.Limit).
			WillReturnRows(orderRows(string(domain.OrderStatusPending)))
		mock.ExpectQuery(`SELECT \* FROM "order_items" WHERE "order_items"\."order_id" = \$1`).
			WithArgs("order-1").
			WillReturnRows(itemRows())
		orders, errorValue := repository.FindByUserID(context.Background(), "user-1", pageRequest)
		if errorValue != nil || len(orders) != 1 || len(orders[0].Items) != 1 {
			t.Fatalf("FindByUserID() = %#v, %v", orders, errorValue)
		}
	})

	t.Run("all", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		mock.ExpectQuery(`SELECT \* FROM "orders" WHERE 1 = 1 ORDER BY created_at DESC LIMIT \$1`).
			WithArgs(pageRequest.Limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "total_amount_cents", "status", "created_at", "updated_at"}))
		orders, errorValue := repository.FindAll(context.Background(), pageRequest)
		if errorValue != nil || len(orders) != 0 {
			t.Fatalf("FindAll() = %#v, %v", orders, errorValue)
		}
	})

	t.Run("query error", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		expectedError := errors.New("query failed")
		mock.ExpectQuery(`SELECT \* FROM "orders"`).WillReturnError(expectedError)
		if _, errorValue := repository.FindAll(context.Background(), pageRequest); !errors.Is(errorValue, expectedError) {
			t.Fatalf("FindAll() error = %v", errorValue)
		}
	})

	t.Run("invalid stored entity", func(t *testing.T) {
		databaseConnection, mock := newMockDatabase(t)
		repository := NewOrderRepository(databaseConnection)
		mock.ExpectQuery(`SELECT \* FROM "orders"`).WillReturnRows(orderRows("INVALID"))
		mock.ExpectQuery(`SELECT \* FROM "order_items"`).WillReturnRows(itemRows())
		if _, errorValue := repository.FindAll(context.Background(), pageRequest); !errors.Is(errorValue, domain.ErrInvalidOrderStatus) {
			t.Fatalf("FindAll() error = %v", errorValue)
		}
	})
}
