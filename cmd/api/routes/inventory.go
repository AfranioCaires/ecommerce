package routes

import (
	"github.com/gin-gonic/gin"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

func RegisterInventoryRoutes(
	apiRoutes *gin.RouterGroup,
	inventoryHandler *inventorytransport.Handler,
	accessTokenParser middleware.AccessTokenParser,
) {
	inventoryRoutes := apiRoutes.Group("/inventory")
	inventoryRoutes.PUT("/:productID",
		middleware.RequireAuthentication(accessTokenParser),
		middleware.RequireAnyRole(string(authenticationdomain.RoleAdministrator)),
		inventoryHandler.SetQuantity,
	)
}
