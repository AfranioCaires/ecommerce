package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type accessTokenParserStub struct {
	claims *security.AccessTokenClaims
	err    error
}

func (accessTokenParserStub) Parse(string) (*security.AccessTokenClaims, error) {
	return &security.AccessTokenClaims{
		Roles: []string{"CUSTOMER"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-1",
		},
	}, nil
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiRoutes := router.Group("/api")

	RegisterAuthenticationRoutes(apiRoutes, authenticationtransport.NewHandler(nil, nil))
	RegisterCatalogRoutes(apiRoutes, catalogtransport.NewHandler(nil, nil, nil), accessTokenParserStub{})
	RegisterInventoryRoutes(apiRoutes, inventorytransport.NewHandler(nil), accessTokenParserStub{})
	RegisterOrderRoutes(
		apiRoutes,
		checkouttransport.NewHandler(nil),
		ordertransport.NewHandler(nil, nil, nil),
		accessTokenParserStub{},
	)

	want := map[string]bool{
		"POST /api/authentication/register": false,
		"POST /api/authentication/login":    false,
		"GET /api/products":                 false,
		"GET /api/products/:productID":      false,
		"POST /api/products":                false,
		"PUT /api/inventory/:productID":     false,
		"POST /api/orders":                  false,
		"GET /api/orders":                   false,
		"GET /api/orders/:orderID":          false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, available := want[key]; available {
			want[key] = true
		}
	}
	for route, available := range want {
		if !available {
			t.Errorf("route %s was not registered", route)
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
