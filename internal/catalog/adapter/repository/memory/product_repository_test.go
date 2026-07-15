package catalogrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
)

func TestProductRepository(t *testing.T) {
	t.Run("it should save and find a product", func(t *testing.T) {
		repository := NewProductRepository()
		product, _ := domain.NewProduct("product-1", "Keyboard", "Mechanical", 10000, time.Now())
		if errorValue := repository.Save(context.Background(), product); errorValue != nil {
			t.Fatal(errorValue)
		}
		storedProduct, errorValue := repository.FindByID(context.Background(), product.ID)
		if errorValue != nil || storedProduct.Name != product.Name {
			t.Fatalf("expected stored product, received %#v and %v", storedProduct, errorValue)
		}
	})

	t.Run("it should return product not found", func(t *testing.T) {
		repository := NewProductRepository()
		_, errorValue := repository.FindByID(context.Background(), "missing")
		if !errors.Is(errorValue, domain.ErrProductNotFound) {
			t.Fatalf("expected product not found, received %v", errorValue)
		}
	})

	t.Run("it should paginate active products", func(t *testing.T) {
		repository := NewProductRepository()
		createdAt := time.Now()
		firstProduct, _ := domain.NewProduct("product-1", "First", "", 100, createdAt)
		secondProduct, _ := domain.NewProduct("product-2", "Second", "", 200, createdAt.Add(time.Second))
		repository.Save(context.Background(), firstProduct)
		repository.Save(context.Background(), secondProduct)
		pageRequest, _ := usecase.NewProductPageRequest(1, 1)
		products, totalItems, errorValue := repository.FindPage(context.Background(), pageRequest)
		if errorValue != nil || totalItems != 2 || len(products) != 1 || products[0].ID != secondProduct.ID {
			t.Fatalf("unexpected page: %#v, %d, %v", products, totalItems, errorValue)
		}
	})

	t.Run("it should find all available requested products", func(t *testing.T) {
		repository := NewProductRepository()
		firstProduct, _ := domain.NewProduct("product-1", "First", "", 100, time.Now())
		secondProduct, _ := domain.NewProduct("product-2", "Second", "", 200, time.Now())
		repository.Save(context.Background(), firstProduct)
		repository.Save(context.Background(), secondProduct)

		products, errorValue := repository.FindByIDs(context.Background(), []string{"product-2", "missing", "product-1"})
		if errorValue != nil || len(products) != 2 || products[0].ID != "product-2" || products[1].ID != "product-1" {
			t.Fatalf("unexpected products: %#v, %v", products, errorValue)
		}
	})

	t.Run("it should sort equal timestamps by ID, ignore inactive products and return an empty page", func(t *testing.T) {
		repository := NewProductRepository()
		createdAt := time.Now()
		firstProduct, _ := domain.NewProduct("product-a", "First", "", 100, createdAt)
		secondProduct, _ := domain.NewProduct("product-b", "Second", "", 200, createdAt)
		inactiveProduct, _ := domain.RestoreProduct("product-c", "Inactive", "", 300, domain.ProductStatusInactive, createdAt.Add(time.Second), createdAt.Add(time.Second))
		repository.Save(context.Background(), secondProduct)
		repository.Save(context.Background(), firstProduct)
		repository.Save(context.Background(), inactiveProduct)

		firstPage, _ := usecase.NewProductPageRequest(1, 10)
		products, totalItems, errorValue := repository.FindPage(context.Background(), firstPage)
		if errorValue != nil || totalItems != 2 || len(products) != 2 || products[0].ID != "product-a" {
			t.Fatalf("unexpected sorted page: %#v, %d, %v", products, totalItems, errorValue)
		}

		emptyPage, _ := usecase.NewProductPageRequest(2, 10)
		products, totalItems, errorValue = repository.FindPage(context.Background(), emptyPage)
		if errorValue != nil || totalItems != 2 || len(products) != 0 {
			t.Fatalf("unexpected empty page: %#v, %d, %v", products, totalItems, errorValue)
		}
	})
}
