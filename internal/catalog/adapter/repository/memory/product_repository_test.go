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
}
