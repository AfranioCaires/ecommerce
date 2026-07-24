package inventorytransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/afraniocaires/ecommerce/internal/inventory/adapter/http/dto"
	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
)

type Handler struct {
	inventoryService *usecase.InventoryService
}

func NewHandler(inventoryService *usecase.InventoryService) *Handler {
	return &Handler{inventoryService: inventoryService}
}

func (handler *Handler) SetQuantity(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	var stockRequest dto.SetStockRequest
	if errorValue := httpresponse.DecodeJSON(responseWriter, request, &stockRequest); errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	stock, errorValue := handler.inventoryService.SetQuantity(
		request.Context(),
		request.PathValue("productID"),
		stockRequest.Quantity,
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		if errors.Is(errorValue, domain.ErrEmptyStockProductID) ||
			errors.Is(errorValue, domain.ErrInvalidStockQuantity) {
			statusCode = http.StatusBadRequest
			message = errorValue.Error()
		}
		httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
		return
	}

	httpresponse.JSON(responseWriter, http.StatusOK, dto.StockResponse{
		ProductID: stock.ProductID,
		Quantity:  stock.Quantity,
		UpdatedAt: stock.UpdatedAt.Format(time.RFC3339),
	})
}
