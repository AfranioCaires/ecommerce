package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type rejectingAccessTokenParser struct{}

func (parser rejectingAccessTokenParser) Parse(accessTokenValue string) (*security.AccessTokenClaims, error) {
	return nil, security.ErrInvalidAccessToken
}

func TestRouter(t *testing.T) {
	router := newRouter(nil, nil, nil, nil, nil, rejectingAccessTokenParser{})

	t.Run("it should expose health", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
		if responseRecorder.Code != http.StatusOK || responseRecorder.Body.String() != "{\"status\":\"UP\"}" {
			t.Fatalf("unexpected health response: %d, %s", responseRecorder.Code, responseRecorder.Body.String())
		}
	})

	t.Run("it should expose swagger documentation", func(t *testing.T) {
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected success, received %d", responseRecorder.Code)
		}
	})

	for description, request := range map[string]*http.Request{
		"it should protect product creation":  httptest.NewRequest(http.MethodPost, "/api/products", nil),
		"it should protect inventory updates": httptest.NewRequest(http.MethodPut, "/api/inventory/product-1", nil),
		"it should protect checkout":          httptest.NewRequest(http.MethodPost, "/api/orders", nil),
		"it should protect order listing":     httptest.NewRequest(http.MethodGet, "/api/orders", nil),
	} {
		t.Run(description, func(t *testing.T) {
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)
			if responseRecorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
			}
		})
	}
}
