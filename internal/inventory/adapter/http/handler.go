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
