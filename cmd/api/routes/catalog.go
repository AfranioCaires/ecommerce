package routes

import (
	"net/http"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterCatalogRoutes(
	router *http.ServeMux,
	productHandler *catalogtransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	router.HandleFunc("GET /api/products", productHandler.List)
	router.HandleFunc("GET /api/products/{productID}", productHandler.GetByID)
	router.Handle(
		"POST /api/products",
		middleware.Chain(
			http.HandlerFunc(productHandler.Create),
			middleware.RequireAuthentication(accessTokenParser),
			middleware.RequireAnyRole(string(authenticationdomain.RoleAdministrator)),
		),
	)
}
