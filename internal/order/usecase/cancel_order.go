package usecase

import (
	"context"
	"errors"
	"time"

	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

var ErrOrderForbidden = errors.New("the order does not belong to the authenticated customer.")

type CancelOrderInput struct {
	OrderID string
	UserID  string
}

type CancelOrderUseCase struct {
	orders      OrderRepository
	inventory   InventoryManager
	transaction TransactionManager
	currentTime func() time.Time
}

func NewCancelOrderUseCase(orders OrderRepository, inventory InventoryManager, transaction TransactionManager, currentTime func() time.Time) *CancelOrderUseCase {
	return &CancelOrderUseCase{orders: orders, inventory: inventory, transaction: transaction, currentTime: currentTime}
}

func (useCase *CancelOrderUseCase) Execute(applicationContext context.Context, input CancelOrderInput) (*domain.Order, error) {
	var canceledOrder *domain.Order
	errorValue := useCase.transaction.Execute(applicationContext, func(transactionContext context.Context) error {
		order, errorValue := useCase.orders.FindByIDForUpdate(transactionContext, input.OrderID)
		if errorValue != nil {
			return errorValue
		}
		if input.UserID != "" && order.UserID != input.UserID {
			return ErrOrderForbidden
		}
		if order.Status != domain.OrderStatusPending {
			return domain.ErrInvalidOrderTransition
		}

		items := make([]inventoryusecase.StockItem, 0, len(order.Items))
		for _, item := range order.Items {
			items = append(items, inventoryusecase.StockItem{ProductID: item.ProductID, Quantity: item.Quantity})
		}
		if errorValue := useCase.inventory.Release(transactionContext, items); errorValue != nil {
			return errorValue
		}
		if errorValue := order.Cancel(useCase.currentTime()); errorValue != nil {
			return errorValue
		}
		if errorValue := useCase.orders.UpdateStatus(transactionContext, order); errorValue != nil {
			return errorValue
		}
		canceledOrder = order
		return nil
	})
	return canceledOrder, errorValue
}
