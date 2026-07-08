package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

var (
	ErrMissingAccessToken = errors.New("the access token is missing.")
	ErrForbidden          = errors.New("access is forbidden.")
	ErrMissingIdentity    = errors.New("the authenticated identity is missing.")
)

const (
	authenticatedUserIDKey = "authenticated_user_id"
	authenticatedRolesKey  = "authenticated_roles"
)

type AccessTokenParser interface {
	Parse(accessTokenValue string) (*security.AccessTokenClaims, error)
}

func RequireAuthentication(accessTokenParser AccessTokenParser) gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationHeader := context.GetHeader("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			context.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpresponse.ErrorResponse{Error: ErrMissingAccessToken.Error()},
			)
			return
		}

		accessTokenValue := strings.TrimSpace(
			strings.TrimPrefix(authorizationHeader, "Bearer "),
		)

		accessTokenClaims, errorValue := accessTokenParser.Parse(accessTokenValue)
		if errorValue != nil || accessTokenClaims.Subject == "" {
			context.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpresponse.ErrorResponse{Error: security.ErrInvalidAccessToken.Error()},
			)
			return
		}

		context.Set(authenticatedUserIDKey, accessTokenClaims.Subject)
		context.Set(authenticatedRolesKey, accessTokenClaims.Roles)
		context.Next()
	}
}

func RequireAnyRole(requiredRoles ...string) gin.HandlerFunc {
	return func(context *gin.Context) {
		authenticatedRoles, available := Roles(context)
		if !available {
			context.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpresponse.ErrorResponse{Error: ErrMissingIdentity.Error()},
			)
			return
		}

		for _, authenticatedRole := range authenticatedRoles {
			for _, requiredRole := range requiredRoles {
				if authenticatedRole == requiredRole {
					context.Next()
					return
				}
			}
		}

		context.AbortWithStatusJSON(
			http.StatusForbidden,
			httpresponse.ErrorResponse{Error: ErrForbidden.Error()},
		)
	}
}

func UserID(context *gin.Context) (string, bool) {
	value, exists := context.Get(authenticatedUserIDKey)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)
	return userID, ok
}

func Roles(context *gin.Context) ([]string, bool) {
	value, exists := context.Get(authenticatedRolesKey)
	if !exists {
		return nil, false
	}

	roles, ok := value.([]string)
	return roles, ok
}
