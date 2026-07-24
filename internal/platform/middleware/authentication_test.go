package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

func TestAuthenticationAndAuthorization(t *testing.T) {
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	accessToken, _ := accessTokenManager.Generate(
		"user-1",
		[]domain.Role{domain.RoleCustomer},
		time.Now(),
	)

	successHandler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		userID, userAvailable := UserID(request.Context())
		roles, rolesAvailable := Roles(request.Context())
		if !userAvailable || !rolesAvailable || userID != "user-1" || len(roles) != 1 {
			http.Error(responseWriter, "missing identity", http.StatusInternalServerError)
			return
		}
		responseWriter.WriteHeader(http.StatusNoContent)
	})

	for _, testCase := range []struct {
		name       string
		token      string
		middleware Middleware
		wantStatus int
	}{
		{name: "missing token", middleware: RequireAuthentication(accessTokenManager), wantStatus: http.StatusUnauthorized},
		{name: "invalid token", token: "invalid", middleware: RequireAuthentication(accessTokenManager), wantStatus: http.StatusUnauthorized},
		{name: "authenticated customer", token: accessToken, middleware: RequireAuthentication(accessTokenManager), wantStatus: http.StatusNoContent},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}
			responseRecorder := httptest.NewRecorder()
			testCase.middleware(successHandler).ServeHTTP(responseRecorder, request)
			if responseRecorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, body = %s", responseRecorder.Code, responseRecorder.Body.String())
			}
		})
	}

	authenticatedHandler := RequireAuthentication(accessTokenManager)(
		RequireAnyRole(string(domain.RoleAdministrator))(successHandler),
	)
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	responseRecorder := httptest.NewRecorder()
	authenticatedHandler.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusForbidden {
		t.Fatalf("role check status = %d", responseRecorder.Code)
	}

	responseRecorder = httptest.NewRecorder()
	RequireAnyRole(string(domain.RoleCustomer))(successHandler).ServeHTTP(
		responseRecorder,
		httptest.NewRequest(http.MethodGet, "/protected", nil),
	)
	if responseRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("missing identity status = %d", responseRecorder.Code)
	}
}

func TestIdentityAccessorsAndChain(t *testing.T) {
	applicationContext := context.Background()
	if _, available := UserID(applicationContext); available {
		t.Fatal("UserID() unexpectedly found an identity")
	}
	if _, available := Roles(applicationContext); available {
		t.Fatal("Roles() unexpectedly found roles")
	}

	sequence := make([]string, 0, 3)
	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			sequence = append(sequence, "first")
			next.ServeHTTP(writer, request)
		})
	}
	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			sequence = append(sequence, "second")
			next.ServeHTTP(writer, request)
		})
	}
	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		sequence = append(sequence, "handler")
	})

	Chain(final, first, second).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)
	if len(sequence) != 3 || sequence[0] != "first" || sequence[1] != "second" || sequence[2] != "handler" {
		t.Fatalf("middleware sequence = %#v", sequence)
	}
}

func TestRecover(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})).ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if responseRecorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", responseRecorder.Code)
	}
}
