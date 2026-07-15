package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
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

type productRepositoryStub struct {
	save     func(*domain.Product) error
	findPage func(usecase.ProductPageRequest) ([]*domain.Product, int64, error)
}

func (stub productRepositoryStub) Save(_ context.Context, product *domain.Product) error {
	return stub.save(product)
}

func (stub productRepositoryStub) FindByID(context.Context, string) (*domain.Product, error) {
	return nil, domain.ErrProductNotFound
}

func (stub productRepositoryStub) FindByIDs(context.Context, []string) ([]*domain.Product, error) {
	return nil, nil
}

func (stub productRepositoryStub) FindPage(_ context.Context, request usecase.ProductPageRequest) ([]*domain.Product, int64, error) {
	return stub.findPage(request)
}

func TestCatalogUseCaseErrorsAndPagination(t *testing.T) {
	expectedError := errors.New("repository failed")
	currentTime := func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }

	t.Run("it should reject invalid product input", func(t *testing.T) {
		useCase := usecase.NewCreateProductUseCase(productRepositoryStub{}, currentTime)
		if _, errorValue := useCase.Execute(context.Background(), usecase.CreateProductInput{Name: " ", PriceCents: 100}); !errors.Is(errorValue, domain.ErrEmptyProductName) {
			t.Fatalf("expected empty name, received %v", errorValue)
		}
	})

	t.Run("it should return a product save error", func(t *testing.T) {
		useCase := usecase.NewCreateProductUseCase(productRepositoryStub{save: func(*domain.Product) error { return expectedError }}, currentTime)
		if _, errorValue := useCase.Execute(context.Background(), usecase.CreateProductInput{Name: "Product", PriceCents: 100}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected save error, received %v", errorValue)
		}
	})

	t.Run("it should return a list error", func(t *testing.T) {
		useCase := usecase.NewListProductsUseCase(productRepositoryStub{findPage: func(usecase.ProductPageRequest) ([]*domain.Product, int64, error) {
			return nil, 0, expectedError
		}})
		if _, errorValue := useCase.Execute(context.Background(), usecase.ProductPageRequest{PageNumber: 1, PageSize: 20}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected list error, received %v", errorValue)
		}
	})

	t.Run("it should validate all invalid pagination forms", func(t *testing.T) {
		for _, request := range []struct{ page, size int }{{0, 20}, {1, 0}, {1, usecase.MaximumPageSize + 1}} {
			if _, errorValue := usecase.NewProductPageRequest(request.page, request.size); !errors.Is(errorValue, usecase.ErrInvalidPagination) {
				t.Fatalf("expected invalid pagination for %#v, received %v", request, errorValue)
			}
		}
	})

	t.Run("it should create an empty product page", func(t *testing.T) {
		page := usecase.NewProductPage(nil, usecase.ProductPageRequest{PageNumber: 1, PageSize: 20}, 0)
		if page.TotalPages != 0 || page.TotalItems != 0 {
			t.Fatalf("unexpected empty page: %#v", page)
		}
	})
}
