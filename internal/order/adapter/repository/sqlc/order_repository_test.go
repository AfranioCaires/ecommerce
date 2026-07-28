package orderrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/order/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

func newOrderRepository(
	t *testing.T,
) (*sql.DB, *OrderRepository, sqlmock.Sqlmock) {
	t.Helper()
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	return databaseConnection, NewOrderRepository(
		databasequeries.New(databaseConnection),
	), mock
}

func orderRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"customer_id",
		"total_amount_cents",
		"status",
		"created_at",
		"updated_at",
	})
}

func orderItemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"order_id",
		"product_id",
		"product_name",
		"unit_price_cents",
		"quantity",
	})
}

func newOrder(t *testing.T) *domain.Order {
	t.Helper()
	order, errorValue := domain.NewOrder(
		"order-1",
		"customer-1",
		[]domain.OrderItem{{
			ProductID:      "product-1",
			ProductName:    "Keyboard",
			UnitPriceCents: 10000,
			Quantity:       2,
		}},
		time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC),
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}

	return order
}

func TestOrderRepositoryCreatesOrderAndItemsInsideTransaction(t *testing.T) {
	databaseConnection, repository, mock := newOrderRepository(t)
	order := newOrder(t)
	manager := transaction.NewManager(databaseConnection)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").
		WithArgs(
			order.ID,
			order.UserID,
			order.TotalAmountCents,
			"PENDING",
			order.CreatedAt,
			order.UpdatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO order_items").
		WithArgs(
			order.ID,
			"product-1",
			"Keyboard",
			int64(10000),
			int32(2),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	errorValue := manager.Execute(context.Background(), func(transactionContext context.Context) error {
		return repository.Save(transactionContext, order)
	})
	if errorValue != nil {
		t.Fatalf("Execute() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestOrderRepositoryFailureRollsBackCreation(t *testing.T) {
	databaseConnection, repository, mock := newOrderRepository(t)
	order := newOrder(t)
	manager := transaction.NewManager(databaseConnection)
	expectedError := errors.New("item insert failed")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO order_items").WillReturnError(expectedError)
	mock.ExpectRollback()

	errorValue := manager.Execute(context.Background(), func(transactionContext context.Context) error {
		return repository.Save(transactionContext, order)
	})
	if !errors.Is(errorValue, expectedError) {
		t.Fatalf("Execute() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestOrderRepositoryFindsOrderWithItems(t *testing.T) {
	_, repository, mock := newOrderRepository(t)
	createdAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("FROM orders").
		WithArgs("order-1").
		WillReturnRows(orderRows().AddRow(
			"order-1",
			"customer-1",
			int64(20000),
			"PAID",
			createdAt,
			createdAt,
		))
	mock.ExpectQuery("FROM order_items").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(orderItemRows().AddRow(
			int64(1),
			"order-1",
			"product-1",
			"Keyboard",
			int64(10000),
			int32(2),
		))

	order, errorValue := repository.FindByID(context.Background(), "order-1")
	if errorValue != nil || order.Status != domain.OrderStatusPaid || len(order.Items) != 1 {
		t.Fatalf("FindByID() = %#v, %v", order, errorValue)
	}

	mock.ExpectQuery("FROM orders").WithArgs("missing").WillReturnError(sql.ErrNoRows)
	if _, errorValue := repository.FindByID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrOrderNotFound) {
		t.Fatalf("FindByID() error = %v", errorValue)
	}
}

func TestOrderRepositoryListsCustomerOrdersWithLimitAndOffset(t *testing.T) {
	_, repository, mock := newOrderRepository(t)
	createdAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("WHERE customer_id").
		WithArgs("customer-1", int32(5), int32(10)).
		WillReturnRows(orderRows().AddRow(
			"order-1",
			"customer-1",
			int64(20000),
			"PAID",
			createdAt,
			createdAt,
		))
	mock.ExpectQuery("FROM order_items").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(orderItemRows().AddRow(
			int64(1),
			"order-1",
			"product-1",
			"Keyboard",
			int64(10000),
			int32(2),
		))

	orders, errorValue := repository.FindByUserID(
		context.Background(),
		"customer-1",
		usecase.OrderPageRequest{Limit: 10, Offset: 5},
	)
	if errorValue != nil || len(orders) != 1 || orders[0].UserID != "customer-1" {
		t.Fatalf("FindByUserID() = %#v, %v", orders, errorValue)
	}
}

func TestOrderRepositoryReturnsEmptyPageWithoutItemQuery(t *testing.T) {
	_, repository, mock := newOrderRepository(t)
	mock.ExpectQuery("FROM orders").
		WithArgs(int32(0), int32(20)).
		WillReturnRows(orderRows())

	orders, errorValue := repository.FindAll(
		context.Background(),
		usecase.OrderPageRequest{Limit: 20, Offset: 0},
	)
	if errorValue != nil || len(orders) != 0 {
		t.Fatalf("FindAll() = %#v, %v", orders, errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}
