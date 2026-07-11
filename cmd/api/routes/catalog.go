package routes

import (
	"github.com/gin-gonic/gin"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterCatalogRoutes(
	apiRoutes *gin.RouterGroup,
	productHandler *catalogtransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	productRoutes := apiRoutes.Group("/products")
	productRoutes.GET("", productHandler.List)
	productRoutes.GET("/:productID", productHandler.GetByID)
	productRoutes.POST("",
		middleware.RequireAuthentication(accessTokenParser),
		middleware.RequireAnyRole(string(authenticationdomain.RoleAdministrator)),
		productHandler.Create,
	)
}
