package ordertransport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type orderRepository struct {
	orders      []*orderdomain.Order
	pageRequest orderusecase.OrderPageRequest
	findError   error
	method      string
}

func (repository *orderRepository) Save(context.Context, *orderdomain.Order) error { return nil }
func (repository *orderRepository) UpdateStatus(context.Context, *orderdomain.Order) error {
	return nil
}
func (repository *orderRepository) FindByID(_ context.Context, orderID string) (*orderdomain.Order, error) {
	repository.method = "FindByID"
	if repository.findError != nil {
		return nil, repository.findError
	}
	for _, order := range repository.orders {
		if order.ID == orderID {
			return order, nil
		}
	}
	return nil, orderdomain.ErrOrderNotFound
}
func (repository *orderRepository) FindByUserID(_ context.Context, userID string, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	repository.method, repository.pageRequest = "FindByUserID", pageRequest
	if repository.findError != nil {
		return nil, repository.findError
	}
	orders := make([]*orderdomain.Order, 0)
	for _, order := range repository.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
func (repository *orderRepository) FindAll(_ context.Context, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	repository.method, repository.pageRequest = "FindAll", pageRequest
	if repository.findError != nil {
		return nil, repository.findError
	}
	return repository.orders, nil
}

func orderRouter(repository *orderRepository, userID string, roles []authenticationdomain.Role, authenticated bool) (http.Handler, string) {
	handler := NewHandler(
		orderusecase.NewGetOrderUseCase(repository),
		orderusecase.NewListUserOrdersUseCase(repository),
		orderusecase.NewListAllOrdersUseCase(repository),
	)
	router := http.NewServeMux()
	var listHandler http.Handler = http.HandlerFunc(handler.List)
	var getHandler http.Handler = http.HandlerFunc(handler.GetByID)
	token := ""
	if authenticated {
		manager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
		token, _ = manager.Generate(userID, roles, time.Now())
		listHandler = middleware.RequireAuthentication(manager)(listHandler)
		getHandler = middleware.RequireAuthentication(manager)(getHandler)
	}
	router.Handle("GET /orders", listHandler)
	router.Handle("GET /orders/{orderID}", getHandler)
	return router, token
}

func performOrderRequest(router http.Handler, token, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestHandler(t *testing.T) {
	createdAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	order, _ := orderdomain.NewOrder("order-1", "user-1", []orderdomain.OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 100, Quantity: 2}}, createdAt)
	dependencyError := errors.New("dependency failed")

	for _, testCase := range []struct {
		name          string
		path          string
		userID        string
		roles         []authenticationdomain.Role
		authenticated bool
		repository    *orderRepository
		wantStatus    int
		wantMethod    string
	}{
		{name: "get requires identity", path: "/orders/order-1", repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusUnauthorized},
		{name: "get maps not found", path: "/orders/missing", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusNotFound, wantMethod: "FindByID"},
		{name: "get maps repository failure", path: "/orders/order-1", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{findError: dependencyError}, wantStatus: http.StatusInternalServerError, wantMethod: "FindByID"},
		{name: "get forbids another customer", path: "/orders/order-1", userID: "user-2", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusForbidden, wantMethod: "FindByID"},
		{name: "get permits owner", path: "/orders/order-1", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusOK, wantMethod: "FindByID"},
		{name: "get permits support", path: "/orders/order-1", userID: "user-2", roles: []authenticationdomain.Role{authenticationdomain.RoleSupport}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusOK, wantMethod: "FindByID"},
		{name: "list requires identity", path: "/orders", repository: &orderRepository{}, wantStatus: http.StatusUnauthorized},
		{name: "list rejects invalid limit string", path: "/orders?limit=x", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{}, wantStatus: http.StatusBadRequest},
		{name: "list rejects invalid offset string", path: "/orders?offset=x", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{}, wantStatus: http.StatusBadRequest},
		{name: "list rejects invalid values", path: "/orders?limit=101", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{}, wantStatus: http.StatusBadRequest},
		{name: "list owned orders with defaults", path: "/orders", userID: "user-1", roles: []authenticationdomain.Role{authenticationdomain.RoleCustomer}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusOK, wantMethod: "FindByUserID"},
		{name: "list all orders for administrator", path: "/orders?limit=10&offset=5", userID: "admin-1", roles: []authenticationdomain.Role{authenticationdomain.RoleAdministrator}, authenticated: true, repository: &orderRepository{orders: []*orderdomain.Order{order}}, wantStatus: http.StatusOK, wantMethod: "FindAll"},
		{name: "list maps repository failure", path: "/orders", userID: "support-1", roles: []authenticationdomain.Role{authenticationdomain.RoleSupport}, authenticated: true, repository: &orderRepository{findError: dependencyError}, wantStatus: http.StatusInternalServerError, wantMethod: "FindAll"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			router, token := orderRouter(testCase.repository, testCase.userID, testCase.roles, testCase.authenticated)
			response := performOrderRequest(router, token, testCase.path)
			if response.Code != testCase.wantStatus || testCase.repository.method != testCase.wantMethod {
				t.Fatalf("status = %d, body = %s, method = %q", response.Code, response.Body.String(), testCase.repository.method)
			}
			if testCase.name == "list owned orders with defaults" && testCase.repository.pageRequest != (orderusecase.OrderPageRequest{Limit: 20, Offset: 0}) {
				t.Fatalf("page request = %#v", testCase.repository.pageRequest)
			}
			if testCase.name == "list all orders for administrator" && testCase.repository.pageRequest != (orderusecase.OrderPageRequest{Limit: 10, Offset: 5}) {
				t.Fatalf("page request = %#v", testCase.repository.pageRequest)
			}
			if response.Code == http.StatusOK {
				var value any
				if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
					t.Fatalf("invalid response JSON: %v", err)
				}
			}
		})
	}
}

func TestOrderHandlerHelpers(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/?value=7&broken=x", nil)
	if value, err := orderQueryValue(request, "missing", 3); value != 3 || err != nil {
		t.Fatalf("fallback = %d, %v", value, err)
	}
	if value, err := orderQueryValue(request, "value", 3); value != 7 || err != nil {
		t.Fatalf("parsed = %d, %v", value, err)
	}
	if _, err := orderQueryValue(request, "broken", 3); err == nil {
		t.Fatal("expected parse failure")
	}
	if containsAnyRole(nil, "ADMINISTRATOR") || containsAnyRole([]string{"CUSTOMER"}, "ADMINISTRATOR") || !containsAnyRole([]string{"SUPPORT"}, "SUPPORT") {
		t.Fatal("unexpected role matching")
	}
}
