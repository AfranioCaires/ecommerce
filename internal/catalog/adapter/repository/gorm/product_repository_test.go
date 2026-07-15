package catalogrepository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
)

var errProductDatabase = errors.New("product database failed")

func newProductMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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

func productRows(createdAt time.Time, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "description", "price_cents", "status", "created_at", "updated_at"}).
		AddRow("product-1", "Product", "Description", int64(100), status, createdAt, createdAt)
}

func TestProductModelAndMappers(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	product, _ := domain.NewProduct("product-1", " Product ", " Description ", 100, now)
	model := toProductModel(product)
	if (ProductModel{}).TableName() != "products" || model.ID != product.ID || model.Name != "Product" || model.Description != "Description" || model.PriceCents != 100 || model.Status != "ACTIVE" || !model.CreatedAt.Equal(now) || !model.UpdatedAt.Equal(now) {
		t.Fatalf("toProductModel() = %#v", model)
	}
	entity, errorValue := toProductEntity(&model)
	if errorValue != nil || entity.ID != product.ID || entity.Status != domain.ProductStatusActive {
		t.Fatalf("toProductEntity() = %#v, %v", entity, errorValue)
	}
	model.Status = "INVALID"
	if entity, errorValue = toProductEntity(&model); entity != nil || !errors.Is(errorValue, domain.ErrInvalidProductStatus) {
		t.Fatalf("toProductEntity(invalid) = %#v, %v", entity, errorValue)
	}
}

func TestProductRepositorySave(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	product, _ := domain.NewProduct("product-1", "Product", "Description", 100, now)
	for _, testCase := range []struct {
		name      string
		dbError   error
		wantError error
	}{{"success", nil, nil}, {"database error", errProductDatabase, errProductDatabase}} {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newProductMockDatabase(t)
			mock.ExpectBegin()
			expectation := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products"`))
			if testCase.dbError != nil {
				expectation.WillReturnError(testCase.dbError)
				mock.ExpectRollback()
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}
			if errorValue := NewProductRepository(database).Save(context.Background(), product); !errors.Is(errorValue, testCase.wantError) {
				t.Fatalf("Save() error = %v", errorValue)
			}
		})
	}
}

func TestProductRepositoryFindByID(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	query := `SELECT .* FROM "products" WHERE id =`
	testCases := []struct {
		name      string
		rows      *sqlmock.Rows
		dbError   error
		wantError error
	}{
		{"not found", sqlmock.NewRows([]string{"id"}), nil, domain.ErrProductNotFound},
		{"database error", nil, errProductDatabase, errProductDatabase},
		{"invalid product", productRows(now, "INVALID"), nil, domain.ErrInvalidProductStatus},
		{"success", productRows(now, "ACTIVE"), nil, nil},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newProductMockDatabase(t)
			expectation := mock.ExpectQuery(query)
			if testCase.dbError != nil {
				expectation.WillReturnError(testCase.dbError)
			} else {
				expectation.WillReturnRows(testCase.rows)
			}
			product, errorValue := NewProductRepository(database).FindByID(context.Background(), "product-1")
			if !errors.Is(errorValue, testCase.wantError) || (testCase.wantError == nil && product.ID != "product-1") {
				t.Fatalf("FindByID() = %#v, %v", product, errorValue)
			}
		})
	}
}

func TestProductRepositoryFindByIDs(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	query := `SELECT .* FROM "products" WHERE id IN`
	testCases := []struct {
		name      string
		rows      *sqlmock.Rows
		dbError   error
		wantError error
		wantCount int
	}{
		{"database error", nil, errProductDatabase, errProductDatabase, 0},
		{"invalid product", productRows(now, "INVALID"), nil, domain.ErrInvalidProductStatus, 0},
		{"success", productRows(now, "ACTIVE"), nil, nil, 1},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newProductMockDatabase(t)
			expectation := mock.ExpectQuery(query)
			if testCase.dbError != nil {
				expectation.WillReturnError(testCase.dbError)
			} else {
				expectation.WillReturnRows(testCase.rows)
			}
			products, errorValue := NewProductRepository(database).FindByIDs(context.Background(), []string{"product-1"})
			if !errors.Is(errorValue, testCase.wantError) || len(products) != testCase.wantCount {
				t.Fatalf("FindByIDs() = %#v, %v", products, errorValue)
			}
		})
	}
}

func TestProductRepositoryFindPage(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	pageRequest := usecase.ProductPageRequest{PageNumber: 2, PageSize: 5}
	countQuery := `SELECT count\(\*\) FROM "products" WHERE status =`
	findQuery := `SELECT .* FROM "products" WHERE status = .*ORDER BY created_at DESC LIMIT`

	t.Run("count error", func(t *testing.T) {
		database, mock := newProductMockDatabase(t)
		mock.ExpectQuery(countQuery).WillReturnError(errProductDatabase)
		products, total, errorValue := NewProductRepository(database).FindPage(context.Background(), pageRequest)
		if products != nil || total != 0 || !errors.Is(errorValue, errProductDatabase) {
			t.Fatalf("FindPage() = %#v, %d, %v", products, total, errorValue)
		}
	})

	for _, testCase := range []struct {
		name      string
		rows      *sqlmock.Rows
		dbError   error
		wantError error
		wantCount int
	}{{"find error", nil, errProductDatabase, errProductDatabase, 0}, {"invalid product", productRows(now, "INVALID"), nil, domain.ErrInvalidProductStatus, 0}, {"success", productRows(now, "ACTIVE"), nil, nil, 1}} {
		t.Run(testCase.name, func(t *testing.T) {
			database, mock := newProductMockDatabase(t)
			mock.ExpectQuery(countQuery).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
			expectation := mock.ExpectQuery(findQuery)
			if testCase.dbError != nil {
				expectation.WillReturnError(testCase.dbError)
			} else {
				expectation.WillReturnRows(testCase.rows)
			}
			products, total, errorValue := NewProductRepository(database).FindPage(context.Background(), pageRequest)
			if !errors.Is(errorValue, testCase.wantError) || len(products) != testCase.wantCount || (testCase.wantError == nil && total != 7) {
				t.Fatalf("FindPage() = %#v, %d, %v", products, total, errorValue)
			}
		})
	}
}
