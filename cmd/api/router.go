package main

import (
	"net/http"

	"github.com/afraniocaires/ecommerce/cmd/api/routes"
	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func newRouter(
	authenticationHandler *authenticationtransport.Handler,
	productHandler *catalogtransport.Handler,
	inventoryHandler *inventorytransport.Handler,
	checkoutHandler *checkouttransport.Handler,
	orderHandler *ordertransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("GET /health", httpresponse.Health)

	routes.RegisterAuthenticationRoutes(router, authenticationHandler)
	routes.RegisterCatalogRoutes(router, productHandler, accessTokenParser)
	routes.RegisterInventoryRoutes(router, inventoryHandler, accessTokenParser)
	routes.RegisterOrderRoutes(
		router,
		checkoutHandler,
		orderHandler,
		accessTokenParser,
	)

	return middleware.Recover(middleware.Log(router))
}
