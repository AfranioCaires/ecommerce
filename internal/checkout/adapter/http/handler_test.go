package checkouttransport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	paymentdomain "github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type checkoutDependencies struct {
	product       *catalogdomain.Product
	paymentStatus paymentdomain.PaymentStatus
}

func (dependencies *checkoutDependencies) FindByIDs(applicationContext context.Context, productIDs []string) ([]*catalogdomain.Product, error) {
	return []*catalogdomain.Product{dependencies.product}, nil
}

func (dependencies *checkoutDependencies) Reserve(applicationContext context.Context, stockItems []inventoryusecase.StockItem) error {
	return nil
}

func (dependencies *checkoutDependencies) Release(applicationContext context.Context, stockItems []inventoryusecase.StockItem) error {
	return nil
}

func (dependencies *checkoutDependencies) Save(applicationContext context.Context, order *orderdomain.Order) error {
	return nil
}

func (dependencies *checkoutDependencies) UpdateStatus(applicationContext context.Context, order *orderdomain.Order) error {
	return nil
}

func (dependencies *checkoutDependencies) Process(applicationContext context.Context, orderID string, amountCents int64) (*paymentdomain.Payment, error) {
	return paymentdomain.NewPayment("payment-1", orderID, amountCents, dependencies.paymentStatus, time.Now())
}

func (dependencies *checkoutDependencies) Execute(applicationContext context.Context, operation func(transactionContext context.Context) error) error {
	return operation(applicationContext)
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	product, _ := catalogdomain.NewProduct("product-1", "Keyboard", "", 10000, time.Now())
	dependencies := &checkoutDependencies{product: product, paymentStatus: paymentdomain.PaymentStatusApproved}
	checkoutUseCase := checkoutusecase.NewCheckoutUseCase(dependencies, dependencies, dependencies, dependencies, dependencies, time.Now)
	handler := NewHandler(checkoutUseCase)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	router := gin.New()
	router.Use(middleware.RequireAuthentication(accessTokenManager))
	router.POST("/orders", handler.Checkout)

	t.Run("it should complete an authenticated checkout", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"items":[{"product_id":"product-1","quantity":2}]}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("expected created, received %d and %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	t.Run("it should reject an unauthenticated checkout", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"items":[]}`))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject malformed JSON", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{"))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})
}
