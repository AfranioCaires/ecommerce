package catalogtransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/catalog/adapter/http/dto"
	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/memory"
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
}
