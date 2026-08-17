package routes

import (
	"net/http"

	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterOrderRoutes(
	router *http.ServeMux,
	checkoutHandler *checkouttransport.Handler,
	orderHandler *ordertransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	authenticationMiddleware := middleware.RequireAuthentication(accessTokenParser)

	router.Handle(
		"POST /api/orders",
		authenticationMiddleware(http.HandlerFunc(checkoutHandler.Checkout)),
	)
	router.Handle(
		"GET /api/orders",
		authenticationMiddleware(http.HandlerFunc(orderHandler.List)),
	)
	router.Handle(
		"GET /api/orders/{orderID}",
		authenticationMiddleware(http.HandlerFunc(orderHandler.GetByID)),
	)
	router.Handle(
		"POST /api/orders/{orderID}/cancel",
		authenticationMiddleware(http.HandlerFunc(orderHandler.Cancel)),
	)
}
