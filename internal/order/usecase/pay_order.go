package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type PayOrderInput struct{ OrderID, UserID string }
type PayOrderOutput struct {
	SagaID, CorrelationID string
	Order                 *domain.Order
}

type PayOrderUseCase struct {
	orders      OrderRepository
	sagas       SagaRepository
	outbox      OutboxWriter
	transaction TransactionManager
	currentTime func() time.Time
}

func NewPayOrderUseCase(orders OrderRepository, sagas SagaRepository, outboxWriter OutboxWriter, transaction TransactionManager, currentTime func() time.Time) *PayOrderUseCase {
	return &PayOrderUseCase{orders: orders, sagas: sagas, outbox: outboxWriter, transaction: transaction, currentTime: currentTime}
}

func (useCase *PayOrderUseCase) Execute(applicationContext context.Context, input PayOrderInput) (*PayOrderOutput, error) {
	var output *PayOrderOutput
	errorValue := useCase.transaction.Execute(applicationContext, func(transactionContext context.Context) error {
		order, errorValue := useCase.orders.FindByIDForUpdate(transactionContext, input.OrderID)
		if errorValue != nil {
			return errorValue
		}
		if input.UserID != "" && order.UserID != input.UserID {
			return ErrOrderForbidden
		}
		now := useCase.currentTime().UTC()
		if errorValue := order.StartPayment(now); errorValue != nil {
			return errorValue
		}
		sagaID, correlationID := uuid.NewString(), uuid.NewString()
		saga, errorValue := domain.NewSaga(sagaID, order.ID, correlationID, now)
		if errorValue != nil {
			return errorValue
		}
		if errorValue := saga.Start(now); errorValue != nil {
			return errorValue
		}
		envelope, errorValue := events.NewEnvelope(uuid.NewString(), events.PaymentRequestedV1, saga.ID, correlationID, now, events.PaymentRequested{OrderID: order.ID, AmountCents: order.TotalAmountCents})
		if errorValue != nil {
			return errorValue
		}
		body, errorValue := json.Marshal(envelope)
		if errorValue != nil {
			return errorValue
		}
		if errorValue := useCase.orders.UpdateStatus(transactionContext, order); errorValue != nil {
			return errorValue
		}
		if errorValue := useCase.sagas.Save(transactionContext, saga); errorValue != nil {
			return errorValue
		}
		if errorValue := useCase.outbox.Save(transactionContext, outbox.NewMessage(envelope.MessageID, envelope.MessageType, "payment.requested", body, now)); errorValue != nil {
			return errorValue
		}
		output = &PayOrderOutput{SagaID: saga.ID, CorrelationID: correlationID, Order: order}
		return nil
	})
	return output, errorValue
}
