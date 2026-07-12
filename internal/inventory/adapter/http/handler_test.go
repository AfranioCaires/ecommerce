package inventorytransport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/inventory/usecase"
)

type stockRepository struct {
	stock *domain.Stock
}

func (repository *stockRepository) Save(applicationContext context.Context, stock *domain.Stock) error {
	repository.stock = stock
	return nil
}

func (repository *stockRepository) FindByProductID(applicationContext context.Context, productID string) (*domain.Stock, error) {
	if repository.stock == nil {
		return nil, domain.ErrStockNotFound
	}
	return repository.stock, nil
}

func (repository *stockRepository) FindByProductIDForUpdate(applicationContext context.Context, productID string) (*domain.Stock, error) {
	return repository.FindByProductID(applicationContext, productID)
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &stockRepository{}
	inventoryService := usecase.NewInventoryService(repository, time.Now)
	handler := NewHandler(inventoryService)
	router := gin.New()
	router.PUT("/inventory/:productID", handler.SetQuantity)

	t.Run("it should set a product quantity", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/inventory/product-1", bytes.NewBufferString(`{"quantity":5}`))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK || repository.stock == nil || repository.stock.Quantity != 5 {
			t.Fatalf("unexpected response: %d, %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	t.Run("it should reject a negative quantity", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/inventory/product-1", bytes.NewBufferString(`{"quantity":-1}`))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})
}
