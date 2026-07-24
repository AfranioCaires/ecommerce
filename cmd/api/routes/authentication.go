package routes

import (
	"net/http"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
)

func RegisterAuthenticationRoutes(
	router *http.ServeMux,
	authenticationHandler *authenticationtransport.Handler,
) {
	router.HandleFunc(
		"POST /api/authentication/register",
		authenticationHandler.Register,
	)
	router.HandleFunc(
		"POST /api/authentication/login",
		authenticationHandler.Login,
	)
}
