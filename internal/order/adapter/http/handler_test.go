package ordertransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type orderRepository struct {
	orders      []*orderdomain.Order
	pageRequest orderusecase.OrderPageRequest
}

func (repository *orderRepository) Save(applicationContext context.Context, order *orderdomain.Order) error {
	repository.orders = append(repository.orders, order)
	return nil
}
func (repository *orderRepository) UpdateStatus(applicationContext context.Context, order *orderdomain.Order) error {
	return nil
}
func (repository *orderRepository) FindByID(applicationContext context.Context, orderID string) (*orderdomain.Order, error) {
	for _, order := range repository.orders {
		if order.ID == orderID {
			return order, nil
		}
	}
	return nil, orderdomain.ErrOrderNotFound
}
func (repository *orderRepository) FindByUserID(applicationContext context.Context, userID string, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	repository.pageRequest = pageRequest
	orders := make([]*orderdomain.Order, 0)
	for _, order := range repository.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
func (repository *orderRepository) FindAll(applicationContext context.Context, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	repository.pageRequest = pageRequest
	return repository.orders, nil
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order, _ := orderdomain.NewOrder("order-1", "user-1", []orderdomain.OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 100, Quantity: 1}}, time.Now())
	repository := &orderRepository{orders: []*orderdomain.Order{order}}
	handler := NewHandler(orderusecase.NewGetOrderUseCase(repository), orderusecase.NewListUserOrdersUseCase(repository), orderusecase.NewListAllOrdersUseCase(repository))
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	router := gin.New()
	router.Use(middleware.RequireAuthentication(accessTokenManager))
	router.GET("/orders", handler.List)
	router.GET("/orders/:orderID", handler.GetByID)

	t.Run("it should allow a customer to read an owned order", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected success, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject another customer order", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-2", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should apply explicit order pagination", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodGet, "/orders?limit=10&offset=5", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusOK || repository.pageRequest.Limit != 10 || repository.pageRequest.Offset != 5 {
			t.Fatalf("unexpected pagination: %#v, status %d", repository.pageRequest, responseRecorder.Code)
		}
	})

	t.Run("it should reject invalid order pagination", func(t *testing.T) {
		accessToken, _ := accessTokenManager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		request := httptest.NewRequest(http.MethodGet, "/orders?limit=101&offset=-1", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})
}
