package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func Health(context *gin.Context) {
	context.JSON(http.StatusOK, HealthResponse{Status: "UP"})
}
