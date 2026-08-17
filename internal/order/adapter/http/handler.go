package ordertransport

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/order/adapter/http/dto"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
	"github.com/afraniocaires/ecommerce/internal/platform/middleware"
)

type Handler struct {
	getOrderUseCase       *orderusecase.GetOrderUseCase
	listUserOrdersUseCase *orderusecase.ListUserOrdersUseCase
	listAllOrdersUseCase  *orderusecase.ListAllOrdersUseCase
	cancelOrderUseCase    *orderusecase.CancelOrderUseCase
}

func NewHandler(
	getOrderUseCase *orderusecase.GetOrderUseCase,
	listUserOrdersUseCase *orderusecase.ListUserOrdersUseCase,
	listAllOrdersUseCase *orderusecase.ListAllOrdersUseCase,
	cancelOrderUseCase *orderusecase.CancelOrderUseCase,
) *Handler {
	return &Handler{
		getOrderUseCase:       getOrderUseCase,
		listUserOrdersUseCase: listUserOrdersUseCase,
		listAllOrdersUseCase:  listAllOrdersUseCase,
		cancelOrderUseCase:    cancelOrderUseCase,
	}
}

func (handler *Handler) GetByIDPublic(responseWriter http.ResponseWriter, request *http.Request) {
	order, errorValue := handler.getOrderUseCase.Execute(request.Context(), request.PathValue("orderID"))
	if errorValue != nil {
		writeOrderError(responseWriter, errorValue)
		return
	}
	httpresponse.JSON(responseWriter, http.StatusOK, toOrderResponse(order))
}

func (handler *Handler) ListPublic(responseWriter http.ResponseWriter, request *http.Request) {
	limit, errorValue := orderQueryValue(request, "limit", orderusecase.DefaultOrderLimit)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: orderusecase.ErrInvalidOrderPagination.Error()})
		return
	}
	offset, errorValue := orderQueryValue(request, "offset", orderusecase.DefaultOrderOffset)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: orderusecase.ErrInvalidOrderPagination.Error()})
		return
	}
	pageRequest, errorValue := orderusecase.NewOrderPageRequest(limit, offset)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}
	orders, errorValue := handler.listAllOrdersUseCase.Execute(request.Context(), pageRequest)
	if errorValue != nil {
		writeOrderError(responseWriter, errorValue)
		return
	}
	responses := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, toOrderResponse(order))
	}
	httpresponse.JSON(responseWriter, http.StatusOK, responses)
}

func (handler *Handler) Cancel(responseWriter http.ResponseWriter, request *http.Request) {
	userID, available := middleware.UserID(request.Context())
	if !available {
		httpresponse.JSON(responseWriter, http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}
	handler.cancel(responseWriter, request, userID)
}

func (handler *Handler) CancelPublic(responseWriter http.ResponseWriter, request *http.Request) {
	handler.cancel(responseWriter, request, "")
}

func (handler *Handler) cancel(responseWriter http.ResponseWriter, request *http.Request, userID string) {
	order, errorValue := handler.cancelOrderUseCase.Execute(request.Context(), orderusecase.CancelOrderInput{
		OrderID: request.PathValue("orderID"), UserID: userID,
	})
	if errorValue != nil {
		writeOrderError(responseWriter, errorValue)
		return
	}
	httpresponse.JSON(responseWriter, http.StatusOK, toOrderResponse(order))
}

func writeOrderError(responseWriter http.ResponseWriter, errorValue error) {
	statusCode := http.StatusInternalServerError
	message := "an unexpected error occurred."
	switch {
	case errors.Is(errorValue, orderdomain.ErrOrderNotFound):
		statusCode, message = http.StatusNotFound, errorValue.Error()
	case errors.Is(errorValue, orderusecase.ErrOrderForbidden):
		statusCode, message = http.StatusForbidden, errorValue.Error()
	case errors.Is(errorValue, orderdomain.ErrInvalidOrderTransition):
		statusCode, message = http.StatusConflict, errorValue.Error()
	}
	httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
}

func (handler *Handler) GetByID(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	authenticatedUserID, available := middleware.UserID(request.Context())
	if !available {
		httpresponse.JSON(responseWriter, http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	authenticatedRoles, _ := middleware.Roles(request.Context())
	order, errorValue := handler.getOrderUseCase.Execute(
		request.Context(),
		request.PathValue("orderID"),
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		if errors.Is(errorValue, orderdomain.ErrOrderNotFound) {
			statusCode = http.StatusNotFound
			message = errorValue.Error()
		}
		httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
		return
	}

	canReadEveryOrder := containsAnyRole(
		authenticatedRoles,
		string(authenticationdomain.RoleAdministrator),
		string(authenticationdomain.RoleSupport),
	)
	if order.UserID != authenticatedUserID && !canReadEveryOrder {
		httpresponse.JSON(responseWriter, http.StatusForbidden, httpresponse.ErrorResponse{Error: middleware.ErrForbidden.Error()})
		return
	}

	httpresponse.JSON(responseWriter, http.StatusOK, toOrderResponse(order))
}

func (handler *Handler) List(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	authenticatedUserID, available := middleware.UserID(request.Context())
	if !available {
		httpresponse.JSON(responseWriter, http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	authenticatedRoles, _ := middleware.Roles(request.Context())
	limit, errorValue := orderQueryValue(request, "limit", orderusecase.DefaultOrderLimit)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: orderusecase.ErrInvalidOrderPagination.Error()})
		return
	}

	offset, errorValue := orderQueryValue(request, "offset", orderusecase.DefaultOrderOffset)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: orderusecase.ErrInvalidOrderPagination.Error()})
		return
	}

	pageRequest, errorValue := orderusecase.NewOrderPageRequest(limit, offset)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	canReadEveryOrder := containsAnyRole(
		authenticatedRoles,
		string(authenticationdomain.RoleAdministrator),
		string(authenticationdomain.RoleSupport),
	)

	var orders []*orderdomain.Order
	if canReadEveryOrder {
		orders, errorValue = handler.listAllOrdersUseCase.Execute(request.Context(), pageRequest)
	} else {
		orders, errorValue = handler.listUserOrdersUseCase.Execute(
			request.Context(),
			authenticatedUserID,
			pageRequest,
		)
	}
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusInternalServerError, httpresponse.ErrorResponse{Error: "an unexpected error occurred."})
		return
	}

	orderResponses := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		orderResponses = append(orderResponses, toOrderResponse(order))
	}

	httpresponse.JSON(responseWriter, http.StatusOK, orderResponses)
}

func orderQueryValue(request *http.Request, name string, fallback int) (int, error) {
	value := request.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}

	return strconv.Atoi(value)
}

func containsAnyRole(authenticatedRoles []string, requiredRoles ...string) bool {
	for _, authenticatedRole := range authenticatedRoles {
		if slices.Contains(requiredRoles, authenticatedRole) {
			return true
		}
	}

	return false
}

func toOrderResponse(order *orderdomain.Order) dto.OrderResponse {
	orderItemResponses := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, orderItem := range order.Items {
		orderItemResponses = append(orderItemResponses, dto.OrderItemResponse{
			ProductID:      orderItem.ProductID,
			ProductName:    orderItem.ProductName,
			UnitPriceCents: orderItem.UnitPriceCents,
			Quantity:       orderItem.Quantity,
			SubtotalCents:  orderItem.SubtotalCents(),
		})
	}

	return dto.OrderResponse{
		ID:               order.ID,
		UserID:           order.UserID,
		Status:           string(order.Status),
		TotalAmountCents: order.TotalAmountCents,
		Items:            orderItemResponses,
		CreatedAt:        order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        order.UpdatedAt.Format(time.RFC3339),
	}
}
