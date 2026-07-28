package catalogrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
)

func newProductRepository(t *testing.T) (*ProductRepository, sqlmock.Sqlmock) {
	t.Helper()
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	return NewProductRepository(databasequeries.New(databaseConnection)), mock
}

func productRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"name",
		"description",
		"price_cents",
		"status",
		"created_at",
		"updated_at",
	})
}

func TestProductRepositoryWritesAndReadsSQL(t *testing.T) {
	repository, mock := newProductRepository(t)
	now := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)
	product, _ := domain.NewProduct(
		"product-1",
		"Keyboard",
		"Mechanical",
		10000,
		now,
	)

	mock.ExpectExec("INSERT INTO products").
		WithArgs("product-1", "Keyboard", "Mechanical", int64(10000), "ACTIVE", now, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if errorValue := repository.Save(context.Background(), product); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}

	mock.ExpectQuery("FROM products").
		WithArgs("product-1").
		WillReturnRows(productRows().AddRow(
			"product-1",
			"Keyboard",
			"Mechanical",
			int64(10000),
			"ACTIVE",
			now,
			now,
		))
	restored, errorValue := repository.FindByID(context.Background(), "product-1")
	if errorValue != nil || restored.ID != product.ID || restored.PriceCents != 10000 {
		t.Fatalf("FindByID() = %#v, %v", restored, errorValue)
	}

	mock.ExpectQuery("FROM products").WithArgs("missing").WillReturnError(sql.ErrNoRows)
	if _, errorValue := repository.FindByID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrProductNotFound) {
		t.Fatalf("FindByID() error = %v", errorValue)
	}
}

func TestProductRepositoryPaginatesWithLimitAndOffset(t *testing.T) {
	repository, mock := newProductRepository(t)
	now := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(int64(21)),
	)
	mock.ExpectQuery("ORDER BY created_at DESC").
		WithArgs(int32(20), int32(10)).
		WillReturnRows(productRows().AddRow(
			"product-21",
			"Mouse",
			"Wireless",
			int64(5000),
			"ACTIVE",
			now,
			now,
		))

	products, totalItems, errorValue := repository.FindPage(
		context.Background(),
		usecase.ProductPageRequest{PageNumber: 3, PageSize: 10},
	)
	if errorValue != nil || totalItems != 21 || len(products) != 1 || products[0].ID != "product-21" {
		t.Fatalf("FindPage() = %#v, %d, %v", products, totalItems, errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestProductRepositoryListsRequestedIDs(t *testing.T) {
	repository, mock := newProductRepository(t)
	now := time.Now().UTC()
	mock.ExpectQuery("WHERE id = ANY").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(productRows().AddRow(
			"product-1",
			"Keyboard",
			"",
			int64(10000),
			"ACTIVE",
			now,
			now,
		))

	products, errorValue := repository.FindByIDs(
		context.Background(),
		[]string{"product-1"},
	)
	if errorValue != nil || len(products) != 1 {
		t.Fatalf("FindByIDs() = %#v, %v", products, errorValue)
	}
}
