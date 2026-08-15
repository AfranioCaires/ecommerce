package checkouttransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/checkout/adapter/http/dto"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorydomain "github.com/afraniocaires/ecommerce/internal/inventory/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type checkoutDependencies struct {
	products     []*catalogdomain.Product
	findError    error
	reserveError error
}

func (dependencies *checkoutDependencies) FindByID(context.Context, string) (*authenticationdomain.User, error) {
	return &authenticationdomain.User{ID: "user-1"}, nil
}

func (dependencies *checkoutDependencies) FindByIDs(context.Context, []string) ([]*catalogdomain.Product, error) {
	return dependencies.products, dependencies.findError
}
func (dependencies *checkoutDependencies) Reserve(context.Context, []inventoryusecase.StockItem) error {
	return dependencies.reserveError
}
func (dependencies *checkoutDependencies) Release(context.Context, []inventoryusecase.StockItem) error {
	return nil
}
func (dependencies *checkoutDependencies) Save(context.Context, *orderdomain.Order) error {
	return nil
}
func (dependencies *checkoutDependencies) UpdateStatus(context.Context, *orderdomain.Order) error {
	return nil
}
func (dependencies *checkoutDependencies) Execute(applicationContext context.Context, operation func(context.Context) error) error {
	return operation(applicationContext)
}

var fixedCheckoutTime = time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)

func checkoutRouter(dependencies *checkoutDependencies, authenticated bool) (http.Handler, string) {
	useCase := checkoutusecase.NewCheckoutUseCase(dependencies, dependencies, dependencies, dependencies, dependencies, func() time.Time { return fixedCheckoutTime })
	handler := NewHandler(useCase)
	router := http.NewServeMux()
	var checkoutHandler http.Handler = http.HandlerFunc(handler.Checkout)
	token := ""
	if authenticated {
		manager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
		token, _ = manager.Generate("user-1", []authenticationdomain.Role{authenticationdomain.RoleCustomer}, time.Now())
		checkoutHandler = middleware.RequireAuthentication(manager)(checkoutHandler)
	}
	router.Handle("POST /orders", checkoutHandler)
	return router, token
}

func performCheckoutRequest(router http.Handler, token, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestHandler(t *testing.T) {
	product, _ := catalogdomain.NewProduct("product-1", "Keyboard", "Mechanical", 10000, fixedCheckoutTime)
	dependencyError := errors.New("dependency failed")
	validBody := `{"items":[{"product_id":"product-1","quantity":2}]}`

	for _, testCase := range []struct {
		name          string
		body          string
		authenticated bool
		dependencies  *checkoutDependencies
		wantStatus    int
	}{
		{name: "missing identity", body: validBody, dependencies: &checkoutDependencies{}, wantStatus: http.StatusUnauthorized},
		{name: "malformed JSON", body: "{", authenticated: true, dependencies: &checkoutDependencies{}, wantStatus: http.StatusBadRequest},
		{name: "domain validation error", body: `{"items":[]}`, authenticated: true, dependencies: &checkoutDependencies{}, wantStatus: http.StatusBadRequest},
		{name: "product not found", body: validBody, authenticated: true, dependencies: &checkoutDependencies{}, wantStatus: http.StatusNotFound},
		{name: "insufficient stock", body: validBody, authenticated: true, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product}, reserveError: inventorydomain.ErrInsufficientStock}, wantStatus: http.StatusConflict},
		{name: "generic failure", body: validBody, authenticated: true, dependencies: &checkoutDependencies{findError: dependencyError}, wantStatus: http.StatusInternalServerError},
		{name: "pending checkout", body: validBody, authenticated: true, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product}}, wantStatus: http.StatusCreated},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			router, token := checkoutRouter(testCase.dependencies, testCase.authenticated)
			response := performCheckoutRequest(router, token, testCase.body)
			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if response.Code == http.StatusCreated {
				var responseBody dto.CheckoutResponse
				if err := json.Unmarshal(response.Body.Bytes(), &responseBody); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if responseBody.OrderID == "" || responseBody.OrderStatus != string(orderdomain.OrderStatusPending) || responseBody.TotalAmountCents != 20000 || responseBody.CreatedAt != fixedCheckoutTime.Format(time.RFC3339) {
					t.Fatalf("response = %#v", responseBody)
				}
				wantItems := []dto.CheckoutItemResponse{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 10000, Quantity: 2, SubtotalCents: 20000}}
				if len(responseBody.Items) != 1 || responseBody.Items[0] != wantItems[0] {
					t.Fatalf("items = %#v", responseBody.Items)
				}
			}
		})
	}
}
