package ordertransport

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

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
}

func NewHandler(
	getOrderUseCase *orderusecase.GetOrderUseCase,
	listUserOrdersUseCase *orderusecase.ListUserOrdersUseCase,
	listAllOrdersUseCase *orderusecase.ListAllOrdersUseCase,
) *Handler {
	return &Handler{
		getOrderUseCase:       getOrderUseCase,
		listUserOrdersUseCase: listUserOrdersUseCase,
		listAllOrdersUseCase:  listAllOrdersUseCase,
	}
}

// GetByID godoc
// @Summary Get an order
// @Description Returns an owned order or any order for an administrator or support user.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param orderID path string true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 401 {object} httpresponse.ErrorResponse
// @Failure 403 {object} httpresponse.ErrorResponse
// @Failure 404 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /api/orders/{orderID} [get]
func (handler *Handler) GetByID(context *gin.Context) {
	authenticatedUserID, available := middleware.UserID(context)
	if !available {
		context.JSON(http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	authenticatedRoles, _ := middleware.Roles(context)

	order, errorValue := handler.getOrderUseCase.Execute(
		context.Request.Context(),
		context.Param("orderID"),
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(errorValue, orderdomain.ErrOrderNotFound) {
			statusCode = http.StatusNotFound
		}

		context.JSON(statusCode, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	canReadEveryOrder := containsAnyRole(
		authenticatedRoles,
		string(authenticationdomain.RoleAdministrator),
		string(authenticationdomain.RoleSupport),
	)

	if order.UserID != authenticatedUserID && !canReadEveryOrder {
		context.JSON(http.StatusForbidden, httpresponse.ErrorResponse{Error: middleware.ErrForbidden.Error()})
		return
	}

	context.JSON(http.StatusOK, toOrderResponse(order))
}

// List godoc
// @Summary List orders
// @Description Returns owned orders or every order for an administrator or support user.
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.OrderResponse
// @Failure 401 {object} httpresponse.ErrorResponse
// @Failure 500 {object} httpresponse.ErrorResponse
// @Router /api/orders [get]
func (handler *Handler) List(context *gin.Context) {
	authenticatedUserID, available := middleware.UserID(context)
	if !available {
		context.JSON(http.StatusUnauthorized, httpresponse.ErrorResponse{Error: middleware.ErrMissingIdentity.Error()})
		return
	}

	authenticatedRoles, _ := middleware.Roles(context)

	canReadEveryOrder := containsAnyRole(
		authenticatedRoles,
		string(authenticationdomain.RoleAdministrator),
		string(authenticationdomain.RoleSupport),
	)

	var orders []*orderdomain.Order
	var errorValue error

	if canReadEveryOrder {
		orders, errorValue = handler.listAllOrdersUseCase.Execute(
			context.Request.Context(),
		)
	} else {
		orders, errorValue = handler.listUserOrdersUseCase.Execute(
			context.Request.Context(),
			authenticatedUserID,
		)
	}

	if errorValue != nil {
		context.JSON(http.StatusInternalServerError, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	orderResponses := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		orderResponses = append(orderResponses, toOrderResponse(order))
	}

	context.JSON(http.StatusOK, orderResponses)
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
