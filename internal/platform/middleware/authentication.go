package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

var (
	ErrMissingAccessToken = errors.New("the access token is missing.")
	ErrForbidden          = errors.New("access is forbidden.")
	ErrMissingIdentity    = errors.New("the authenticated identity is missing.")
)

type identityContextKey string

const (
	authenticatedUserIDKey identityContextKey = "authenticated_user_id"
	authenticatedRolesKey  identityContextKey = "authenticated_roles"
)

type AccessTokenParser interface {
	Parse(accessTokenValue string) (*security.AccessTokenClaims, error)
}

type Middleware func(http.Handler) http.Handler

func RequireAuthentication(accessTokenParser AccessTokenParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			authorizationHeader := request.Header.Get("Authorization")
			if !strings.HasPrefix(authorizationHeader, "Bearer ") {
				httpresponse.JSON(
					responseWriter,
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
				httpresponse.JSON(
					responseWriter,
					http.StatusUnauthorized,
					httpresponse.ErrorResponse{Error: security.ErrInvalidAccessToken.Error()},
				)
				return
			}

			applicationContext := context.WithValue(
				request.Context(),
				authenticatedUserIDKey,
				accessTokenClaims.Subject,
			)
			applicationContext = context.WithValue(
				applicationContext,
				authenticatedRolesKey,
				accessTokenClaims.Roles,
			)

			next.ServeHTTP(responseWriter, request.WithContext(applicationContext))
		})
	}
}

func RequireAnyRole(requiredRoles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			authenticatedRoles, available := Roles(request.Context())
			if !available {
				httpresponse.JSON(
					responseWriter,
					http.StatusUnauthorized,
					httpresponse.ErrorResponse{Error: ErrMissingIdentity.Error()},
				)
				return
			}

			for _, authenticatedRole := range authenticatedRoles {
				for _, requiredRole := range requiredRoles {
					if authenticatedRole == requiredRole {
						next.ServeHTTP(responseWriter, request)
						return
					}
				}
			}

			httpresponse.JSON(
				responseWriter,
				http.StatusForbidden,
				httpresponse.ErrorResponse{Error: ErrForbidden.Error()},
			)
		})
	}
}

func UserID(applicationContext context.Context) (string, bool) {
	userID, available := applicationContext.Value(authenticatedUserIDKey).(string)
	return userID, available
}

func Roles(applicationContext context.Context) ([]string, bool) {
	roles, available := applicationContext.Value(authenticatedRolesKey).([]string)
	return roles, available
}

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for index := len(middlewares) - 1; index >= 0; index-- {
		handler = middlewares[index](handler)
	}

	return handler
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		defer func() {
			if recoveredValue := recover(); recoveredValue != nil {
				slog.Error(
					"The HTTP handler panicked.",
					"error",
					recoveredValue,
					"stack",
					string(debug.Stack()),
				)
				httpresponse.JSON(
					responseWriter,
					http.StatusInternalServerError,
					httpresponse.ErrorResponse{Error: "an unexpected error occurred."},
				)
			}
		}()

		next.ServeHTTP(responseWriter, request)
	})
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(responseWriter, request)
		slog.Info(
			"HTTP request completed.",
			"method",
			request.Method,
			"path",
			request.URL.Path,
			"duration",
			time.Since(startedAt),
		)
	})
}
