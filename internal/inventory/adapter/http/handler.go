package inventorytransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/inventory/adapter/http/dto"
	"github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
)

type Handler struct {
	inventoryService *usecase.InventoryService
}

func NewHandler(inventoryService *usecase.InventoryService) *Handler {
	return &Handler{inventoryService: inventoryService}
}

// SetQuantity godoc
// @Summary Set product stock
// @Description Replaces the available quantity for a product. Administrator access is required.
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param productID path string true "Product ID"
// @Param request body dto.SetStockRequest true "Stock data"
// @Success 200 {object} dto.StockResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 401 {object} httpresponse.ErrorResponse
// @Failure 403 {object} httpresponse.ErrorResponse
// @Router /api/inventory/{productID} [put]
func (handler *Handler) SetQuantity(context *gin.Context) {
	var request dto.SetStockRequest

	if errorValue := context.ShouldBindJSON(&request); errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	stock, errorValue := handler.inventoryService.SetQuantity(
		context.Request.Context(),
		context.Param("productID"),
		request.Quantity,
	)
	if errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	context.JSON(http.StatusOK, dto.StockResponse{
		ProductID: stock.ProductID,
		Quantity:  stock.Quantity,
		UpdatedAt: stock.UpdatedAt.Format(time.RFC3339),
	})
}
