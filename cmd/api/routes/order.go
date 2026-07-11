package routes

import (
	"github.com/gin-gonic/gin"

	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterOrderRoutes(
	apiRoutes *gin.RouterGroup,
	checkoutHandler *checkouttransport.Handler,
	orderHandler *ordertransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	orderRoutes := apiRoutes.Group("/orders")
	orderRoutes.Use(middleware.RequireAuthentication(accessTokenParser))
	orderRoutes.POST("", checkoutHandler.Checkout)
	orderRoutes.GET("", orderHandler.List)
	orderRoutes.GET("/:orderID", orderHandler.GetByID)
}
