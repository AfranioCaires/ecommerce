package checkouttransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/checkout/adapter/http/dto"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorydomain "github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

type Handler struct {
	checkoutUseCase *checkoutusecase.CheckoutUseCase
}

func NewHandler(
	checkoutUseCase *checkoutusecase.CheckoutUseCase,
) *Handler {
	return &Handler{checkoutUseCase: checkoutUseCase}
}

func (handler *Handler) Checkout(context *gin.Context) {
	userID, available := middleware.UserID(context)
	if !available {
		context.JSON(http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	var request dto.CheckoutRequest

	if errorValue := context.ShouldBindJSON(&request); errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	checkoutItems := make([]checkoutusecase.CheckoutItemInput, 0, len(request.Items))

	for _, requestItem := range request.Items {
		checkoutItems = append(checkoutItems, checkoutusecase.CheckoutItemInput{
			ProductID: requestItem.ProductID,
			Quantity:  requestItem.Quantity,
		})
	}

	output, errorValue := handler.checkoutUseCase.Execute(
		context.Request.Context(),
		checkoutusecase.CheckoutInput{
			UserID: userID,
			Items:  checkoutItems,
		},
	)
	if errorValue != nil {
		statusCode := http.StatusBadRequest

		if errors.Is(errorValue, inventorydomain.ErrInsufficientStock) {
			statusCode = http.StatusConflict
		}
		if errors.Is(errorValue, checkoutusecase.ErrCheckoutProductNotFound) {
			statusCode = http.StatusNotFound
		}

		context.JSON(statusCode, httpresponse.ErrorResponse{Error: errorValue.Error()})
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

	context.JSON(http.StatusCreated, dto.CheckoutResponse{
		OrderID:          output.Order.ID,
		OrderStatus:      string(output.Order.Status),
		PaymentID:        output.Payment.ID,
		PaymentStatus:    string(output.Payment.Status),
		TotalAmountCents: output.Order.TotalAmountCents,
		Items:            itemResponses,
		CreatedAt:        output.Order.CreatedAt.Format(time.RFC3339),
	})
}
