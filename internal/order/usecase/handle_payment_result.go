package usecase

import (
	"context"
	"errors"
	"time"

	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
)

var ErrPaymentResultMismatch = errors.New("the payment result does not match the order saga.")

type HandlePaymentResultInput struct {
	Envelope events.Envelope
	Result   events.PaymentResult
	Approved bool
}

type HandlePaymentResultUseCase struct {
	orders      OrderRepository
	sagas       SagaRepository
	inbox       InboxWriter
	inventory   InventoryManager
	transaction TransactionManager
	currentTime func() time.Time
}

func NewHandlePaymentResultUseCase(orders OrderRepository, sagas SagaRepository, inbox InboxWriter, inventory InventoryManager, transaction TransactionManager, currentTime func() time.Time) *HandlePaymentResultUseCase {
	return &HandlePaymentResultUseCase{orders: orders, sagas: sagas, inbox: inbox, inventory: inventory, transaction: transaction, currentTime: currentTime}
}

func (useCase *HandlePaymentResultUseCase) Execute(applicationContext context.Context, input HandlePaymentResultInput) error {
	return useCase.transaction.Execute(applicationContext, func(transactionContext context.Context) error {
		now := useCase.currentTime().UTC()
		created, errorValue := useCase.inbox.TrySave(transactionContext, input.Envelope.MessageID, now)
		if errorValue != nil || !created {
			return errorValue
		}
		saga, errorValue := useCase.sagas.FindByIDForUpdate(transactionContext, input.Envelope.SagaID)
		if errorValue != nil {
			return errorValue
		}
		order, errorValue := useCase.orders.FindByIDForUpdate(transactionContext, saga.OrderID)
		if errorValue != nil {
			return errorValue
		}
		if saga.CorrelationID != input.Envelope.CorrelationID || order.ID != input.Result.OrderID || order.TotalAmountCents != input.Result.AmountCents || saga.Status != domain.SagaStatusProcessing {
			return ErrPaymentResultMismatch
		}
		if input.Approved {
			if errorValue := order.MarkAsPaid(now); errorValue != nil {
				return errorValue
			}
			if errorValue := saga.Complete(now); errorValue != nil {
				return errorValue
			}
		} else {
			items := make([]inventoryusecase.StockItem, 0, len(order.Items))
			for _, item := range order.Items {
				items = append(items, inventoryusecase.StockItem{ProductID: item.ProductID, Quantity: item.Quantity})
			}
			if errorValue := useCase.inventory.Release(transactionContext, items); errorValue != nil {
				return errorValue
			}
			if errorValue := order.Cancel(now); errorValue != nil {
				return errorValue
			}
			if errorValue := saga.Compensate(now); errorValue != nil {
				return errorValue
			}
		}
		if errorValue := useCase.orders.UpdateStatus(transactionContext, order); errorValue != nil {
			return errorValue
		}
		return useCase.sagas.Update(transactionContext, saga)
	})
}
