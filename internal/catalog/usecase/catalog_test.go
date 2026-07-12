package usecase_test

import (
	"context"
	"testing"
	"time"

	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
)

func TestCatalogUseCases(t *testing.T) {
	productRepository := catalogrepository.NewProductRepository()
	currentTime := func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }
	createProductUseCase := usecase.NewCreateProductUseCase(productRepository, currentTime)
	getProductUseCase := usecase.NewGetProductUseCase(productRepository)
	listProductsUseCase := usecase.NewListProductsUseCase(productRepository)

	t.Run("it should create find and list a product", func(t *testing.T) {
		product, errorValue := createProductUseCase.Execute(context.Background(), usecase.CreateProductInput{Name: "Keyboard", Description: "Mechanical", PriceCents: 10000})
		if errorValue != nil {
			t.Fatal(errorValue)
		}
		storedProduct, errorValue := getProductUseCase.Execute(context.Background(), product.ID)
		if errorValue != nil || storedProduct.ID != product.ID {
			t.Fatalf("unexpected product: %#v, %v", storedProduct, errorValue)
		}
		pageRequest, _ := usecase.NewProductPageRequest(1, 20)
		productPage, errorValue := listProductsUseCase.Execute(context.Background(), pageRequest)
		if errorValue != nil || productPage.TotalItems != 1 || len(productPage.Products) != 1 {
			t.Fatalf("unexpected page: %#v, %v", productPage, errorValue)
		}
	})
}
