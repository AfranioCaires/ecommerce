package httpresponse

import "net/http"

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
func Health(responseWriter http.ResponseWriter, _ *http.Request) {
	JSON(responseWriter, http.StatusOK, HealthResponse{Status: "UP"})
}
