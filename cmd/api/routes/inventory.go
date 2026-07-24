package routes

import (
	"net/http"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterInventoryRoutes(
	router *http.ServeMux,
	inventoryHandler *inventorytransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	router.Handle(
		"PUT /api/inventory/{productID}",
		middleware.Chain(
			http.HandlerFunc(inventoryHandler.SetQuantity),
			middleware.RequireAuthentication(accessTokenParser),
			middleware.RequireAnyRole(string(authenticationdomain.RoleAdministrator)),
		),
	)
}
