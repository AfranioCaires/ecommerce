package routes

import (
	"github.com/gin-gonic/gin"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
)

func RegisterAuthenticationRoutes(
	apiRoutes *gin.RouterGroup,
	authenticationHandler *authenticationtransport.Handler,
) {
	authenticationRoutes := apiRoutes.Group("/authentication")
	authenticationRoutes.POST("/register", authenticationHandler.Register)
	authenticationRoutes.POST("/login", authenticationHandler.Login)
}
