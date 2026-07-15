package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

func TestAuthenticationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	accessToken, _ := accessTokenManager.Generate("user-1", []domain.Role{domain.RoleCustomer}, time.Now())

	t.Run("it should reject a missing access token", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthentication(accessTokenManager))
		router.GET("/protected", func(context *gin.Context) { context.Status(http.StatusNoContent) })
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/protected", nil))
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should expose authenticated identity", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthentication(accessTokenManager), RequireAnyRole(string(domain.RoleCustomer)))
		router.GET("/protected", func(context *gin.Context) {
			userID, available := UserID(context)
			if !available || userID != "user-1" {
				context.Status(http.StatusInternalServerError)
				return
			}
			context.Status(http.StatusNoContent)
		})
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusNoContent {
			t.Fatalf("expected success, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject a role that is not allowed", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthentication(accessTokenManager), RequireAnyRole(string(domain.RoleAdministrator)))
		router.GET("/protected", func(context *gin.Context) { context.Status(http.StatusNoContent) })
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusForbidden {
			t.Fatalf("expected forbidden, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject an invalid access token", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAuthentication(accessTokenManager))
		router.GET("/protected", func(context *gin.Context) { context.Status(http.StatusNoContent) })
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", "Bearer invalid")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject claims without a subject", func(t *testing.T) {
		tokenWithoutSubject, errorValue := accessTokenManager.Generate("", nil, time.Now())
		if errorValue != nil {
			t.Fatal(errorValue)
		}
		router := gin.New()
		router.Use(RequireAuthentication(accessTokenManager))
		router.GET("/protected", func(context *gin.Context) { context.Status(http.StatusNoContent) })
		request := httptest.NewRequest(http.MethodGet, "/protected", nil)
		request.Header.Set("Authorization", "Bearer "+tokenWithoutSubject)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject role checks without an identity", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAnyRole(string(domain.RoleCustomer)))
		router.GET("/protected", func(context *gin.Context) { context.Status(http.StatusNoContent) })
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/protected", nil))
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})
}

func TestIdentityAccessors(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	if userID, available := UserID(context); available || userID != "" {
		t.Fatalf("UserID() = %q, %t", userID, available)
	}
	context.Set(authenticatedUserIDKey, 7)
	if userID, available := UserID(context); available || userID != "" {
		t.Fatalf("UserID() = %q, %t", userID, available)
	}
	context.Set(authenticatedUserIDKey, "user-1")
	if userID, available := UserID(context); !available || userID != "user-1" {
		t.Fatalf("UserID() = %q, %t", userID, available)
	}

	if roles, available := Roles(context); available || roles != nil {
		t.Fatalf("Roles() = %#v, %t", roles, available)
	}
	context.Set(authenticatedRolesKey, "CUSTOMER")
	if roles, available := Roles(context); available || roles != nil {
		t.Fatalf("Roles() = %#v, %t", roles, available)
	}
	context.Set(authenticatedRolesKey, []string{"CUSTOMER"})
	roles, available := Roles(context)
	if !available || len(roles) != 1 || roles[0] != "CUSTOMER" {
		t.Fatalf("Roles() = %#v, %t", roles, available)
	}
}
