package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

// Health godoc
// @Summary Get application health
// @Description Returns the current application health status.
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func Health(context *gin.Context) {
	context.JSON(http.StatusOK, HealthResponse{Status: "UP"})
}
