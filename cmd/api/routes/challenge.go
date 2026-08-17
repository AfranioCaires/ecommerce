package routes

import (
	"net/http"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
)

func RegisterChallengeRoutes(
	router *http.ServeMux,
	authenticationHandler *authenticationtransport.Handler,
	productHandler *catalogtransport.Handler,
	checkoutHandler *checkouttransport.Handler,
	orderHandler *ordertransport.Handler,
) {
	router.HandleFunc("POST /clientes", authenticationHandler.CreateCustomer)
	router.HandleFunc("GET /clientes", authenticationHandler.ListCustomers)
	router.HandleFunc("GET /clientes/{customerID}", authenticationHandler.GetCustomer)
	router.HandleFunc("POST /produtos", productHandler.Create)
	router.HandleFunc("GET /produtos", productHandler.List)
	router.HandleFunc("GET /produtos/{productID}", productHandler.GetByID)

	router.HandleFunc("POST /pedidos", checkoutHandler.CreateForCustomer)
	router.HandleFunc("GET /pedidos", orderHandler.ListPublic)
	router.HandleFunc("GET /pedidos/{orderID}", orderHandler.GetByIDPublic)
	router.HandleFunc("POST /pedidos/{orderID}/cancelar", orderHandler.CancelPublic)
}
