package checkouttransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/afraniocaires/ecommerce/internal/checkout/adapter/http/dto"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorydomain "github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

type Handler struct {
	checkoutUseCase *checkoutusecase.CheckoutUseCase
}

func NewHandler(checkoutUseCase *checkoutusecase.CheckoutUseCase) *Handler {
	return &Handler{checkoutUseCase: checkoutUseCase}
}

func (handler *Handler) Checkout(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	userID, available := middleware.UserID(request.Context())
	if !available {
		httpresponse.JSON(responseWriter, http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	var checkoutRequest dto.CheckoutRequest
	if errorValue := httpresponse.DecodeJSON(responseWriter, request, &checkoutRequest); errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	checkoutItems := make([]checkoutusecase.CheckoutItemInput, 0, len(checkoutRequest.Items))
	for _, requestItem := range checkoutRequest.Items {
		checkoutItems = append(checkoutItems, checkoutusecase.CheckoutItemInput{
			ProductID: requestItem.ProductID,
			Quantity:  requestItem.Quantity,
		})
	}

	output, errorValue := handler.checkoutUseCase.Execute(
		request.Context(),
		checkoutusecase.CheckoutInput{UserID: userID, Items: checkoutItems},
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		switch {
		case errors.Is(errorValue, inventorydomain.ErrInsufficientStock):
			statusCode = http.StatusConflict
			message = errorValue.Error()
		case errors.Is(errorValue, checkoutusecase.ErrCheckoutProductNotFound):
			statusCode = http.StatusNotFound
			message = errorValue.Error()
		case errors.Is(errorValue, checkoutusecase.ErrCheckoutCustomerNotFound):
			statusCode = http.StatusNotFound
			message = errorValue.Error()
		case errors.Is(errorValue, checkoutusecase.ErrEmptyCheckoutItems),
			errors.Is(errorValue, checkoutusecase.ErrInvalidCheckoutItem),
			errors.Is(errorValue, checkoutusecase.ErrInactiveCheckoutProduct):
			statusCode = http.StatusBadRequest
			message = errorValue.Error()
		}

		httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
		return
	}

	itemResponses := make([]dto.CheckoutItemResponse, 0, len(output.Order.Items))
	for _, orderItem := range output.Order.Items {
		itemResponses = append(itemResponses, dto.CheckoutItemResponse{
			ProductID:      orderItem.ProductID,
			ProductName:    orderItem.ProductName,
			UnitPriceCents: orderItem.UnitPriceCents,
			Quantity:       orderItem.Quantity,
			SubtotalCents:  orderItem.SubtotalCents(),
		})
	}

	httpresponse.JSON(responseWriter, http.StatusCreated, dto.CheckoutResponse{
		OrderID:          output.Order.ID,
		OrderStatus:      string(output.Order.Status),
		TotalAmountCents: output.Order.TotalAmountCents,
		Items:            itemResponses,
		CreatedAt:        output.Order.CreatedAt.Format(time.RFC3339),
	})
}
