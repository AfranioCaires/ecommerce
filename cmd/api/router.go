package main

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/afraniocaires/ecommerce/cmd/api/routes"
	_ "github.com/afraniocaires/ecommerce/docs/swagger"
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
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/health", httpresponse.Health)
	router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	apiRoutes := router.Group("/api")
	routes.RegisterAuthenticationRoutes(apiRoutes, authenticationHandler)
	routes.RegisterCatalogRoutes(apiRoutes, productHandler, accessTokenParser)
	routes.RegisterInventoryRoutes(apiRoutes, inventoryHandler, accessTokenParser)
	routes.RegisterOrderRoutes(apiRoutes, checkoutHandler, orderHandler, accessTokenParser)
	return router
}
