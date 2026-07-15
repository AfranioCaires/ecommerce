package catalogtransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/catalog/adapter/http/dto"
	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
)

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	productRepository := catalogrepository.NewProductRepository()
	createProductUseCase := usecase.NewCreateProductUseCase(productRepository, time.Now)
	getProductUseCase := usecase.NewGetProductUseCase(productRepository)
	listProductsUseCase := usecase.NewListProductsUseCase(productRepository)
	handler := NewHandler(createProductUseCase, getProductUseCase, listProductsUseCase)
	router := gin.New()
	router.POST("/products", handler.Create)
	router.GET("/products", handler.List)
	router.GET("/products/:productID", handler.GetByID)

	var productResponse dto.ProductResponse
	t.Run("it should create a product", func(t *testing.T) {
		requestBody, _ := json.Marshal(dto.CreateProductRequest{Name: "Keyboard", Description: "Mechanical", PriceCents: 10000})
		request := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(requestBody))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("expected created, received %d and %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		if errorValue := json.Unmarshal(responseRecorder.Body.Bytes(), &productResponse); errorValue != nil || productResponse.ID == "" {
			t.Fatalf("unexpected response: %#v, %v", productResponse, errorValue)
		}
	})

	t.Run("it should find a product by ID", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products/"+productResponse.ID, nil))
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected success, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should list a product page", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products?page=1&page_size=20", nil))
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected success, received %d", responseRecorder.Code)
		}
		var response dto.ProductPageResponse
		if errorValue := json.Unmarshal(responseRecorder.Body.Bytes(), &response); errorValue != nil || response.TotalItems != 1 {
			t.Fatalf("unexpected response: %#v, %v", response, errorValue)
		}
	})

	t.Run("it should reject invalid pagination", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products?page=0", nil))
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject malformed product JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("{"))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject invalid product data", func(t *testing.T) {
		requestBody, _ := json.Marshal(dto.CreateProductRequest{Name: " ", PriceCents: 100})
		request := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(requestBody))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should return not found for a missing product", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products/missing", nil))
		if responseRecorder.Code != http.StatusNotFound {
			t.Fatalf("expected not found, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject noninteger pagination", func(t *testing.T) {
		for _, path := range []string{"/products?page=invalid", "/products?page_size=invalid"} {
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, path, nil))
			if responseRecorder.Code != http.StatusBadRequest {
				t.Fatalf("expected bad request for %s, received %d", path, responseRecorder.Code)
			}
		}
	})
}

type productRepositoryStub struct {
	findByID func(string) (*domain.Product, error)
	findPage func(usecase.ProductPageRequest) ([]*domain.Product, int64, error)
}

func (stub productRepositoryStub) Save(context.Context, *domain.Product) error { return nil }

func (stub productRepositoryStub) FindByID(_ context.Context, productID string) (*domain.Product, error) {
	return stub.findByID(productID)
}

func (stub productRepositoryStub) FindByIDs(context.Context, []string) ([]*domain.Product, error) {
	return nil, nil
}

func (stub productRepositoryStub) FindPage(_ context.Context, request usecase.ProductPageRequest) ([]*domain.Product, int64, error) {
	return stub.findPage(request)
}

func TestHandlerDependencyErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	expectedError := errors.New("repository failed")
	repository := productRepositoryStub{
		findByID: func(string) (*domain.Product, error) { return nil, expectedError },
		findPage: func(usecase.ProductPageRequest) ([]*domain.Product, int64, error) { return nil, 0, expectedError },
	}
	handler := NewHandler(
		usecase.NewCreateProductUseCase(repository, time.Now),
		usecase.NewGetProductUseCase(repository),
		usecase.NewListProductsUseCase(repository),
	)
	router := gin.New()
	router.GET("/products", handler.List)
	router.GET("/products/:productID", handler.GetByID)

	t.Run("it should return internal server error for product lookup failures", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products/product-1", nil))
		if responseRecorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected internal server error, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should return internal server error for product list failures", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/products", nil))
		if responseRecorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected internal server error, received %d", responseRecorder.Code)
		}
	})
}
