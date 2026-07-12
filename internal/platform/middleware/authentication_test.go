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
}
