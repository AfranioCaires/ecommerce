package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type accessTokenParserStub struct{}

func (accessTokenParserStub) Parse(string) (*security.AccessTokenClaims, error) {
	return &security.AccessTokenClaims{
		Roles: []string{"CUSTOMER"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-1",
		},
	}, nil
}

func TestRegisterRoutes(t *testing.T) {
	router := http.NewServeMux()
	authenticationHandler := authenticationtransport.NewHandler(nil, nil, nil, nil)
	productHandler := catalogtransport.NewHandler(nil, nil, nil)
	RegisterAuthenticationRoutes(router, authenticationHandler)
	RegisterCatalogRoutes(router, productHandler, accessTokenParserStub{})
	RegisterChallengeRoutes(router, authenticationHandler, productHandler, checkouttransport.NewHandler(nil), ordertransport.NewHandler(nil, nil, nil, nil, nil))
	RegisterInventoryRoutes(router, inventorytransport.NewHandler(nil), accessTokenParserStub{})
	RegisterOrderRoutes(
		router,
		checkouttransport.NewHandler(nil),
		ordertransport.NewHandler(nil, nil, nil, nil, nil),
		accessTokenParserStub{},
	)

	for _, route := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/authentication/register"},
		{http.MethodPost, "/api/authentication/login"},
		{http.MethodGet, "/api/products"},
		{http.MethodGet, "/api/products/product-1"},
		{http.MethodPost, "/api/products"},
		{http.MethodPut, "/api/inventory/product-1"},
		{http.MethodPost, "/api/orders"},
		{http.MethodGet, "/api/orders"},
		{http.MethodGet, "/api/orders/order-1"},
	} {
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodOptions, route.path, nil)
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s was not registered: OPTIONS returned %d", route.method, route.path, responseRecorder.Code)
		}
	}

	for _, protectedRoute := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/products"},
		{http.MethodPut, "/api/inventory/product-1"},
		{http.MethodGet, "/api/orders"},
	} {
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(protectedRoute.method, protectedRoute.path, nil)
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without token returned %d", protectedRoute.method, protectedRoute.path, responseRecorder.Code)
		}
	}

	for _, administratorRoute := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/products"},
		{http.MethodPut, "/api/inventory/product-1"},
	} {
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(administratorRoute.method, administratorRoute.path, nil)
		request.Header.Set("Authorization", "Bearer customer-token")
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Errorf("%s %s as customer returned %d", administratorRoute.method, administratorRoute.path, responseRecorder.Code)
		}
	}
}
