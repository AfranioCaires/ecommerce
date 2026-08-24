package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type sagaRepositoryStub struct{ saga *domain.Saga }

func (repository *sagaRepositoryStub) Save(_ context.Context, saga *domain.Saga) error {
	repository.saga = saga
	return nil
}
func (repository *sagaRepositoryStub) FindByIDForUpdate(context.Context, string) (*domain.Saga, error) {
	return repository.saga, nil
}
func (repository *sagaRepositoryStub) Update(_ context.Context, saga *domain.Saga) error {
	repository.saga = saga
	return nil
}

type outboxWriterStub struct{ message *outbox.Message }

func (writer *outboxWriterStub) Save(_ context.Context, message *outbox.Message) error {
	writer.message = message
	return nil
}

type inboxWriterStub struct{ created bool }

func (writer *inboxWriterStub) TrySave(context.Context, string, time.Time) (bool, error) {
	return writer.created, nil
}

func TestPayAndCompleteOrderSaga(t *testing.T) {
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	order, _ := domain.NewOrder("order-1", "customer-1", []domain.OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Quantity: 1}}, now)
	orders := &cancelDependencies{order: order}
	sagas, outboxWriter := &sagaRepositoryStub{}, &outboxWriterStub{}
	pay := NewPayOrderUseCase(orders, sagas, outboxWriter, orders, func() time.Time { return now })
	output, errorValue := pay.Execute(context.Background(), PayOrderInput{OrderID: order.ID, UserID: order.UserID})
	if errorValue != nil || output.Order.Status != domain.OrderStatusPaymentPending || sagas.saga.Status != domain.SagaStatusProcessing || outboxWriter.message.RoutingKey != "payment.requested" {
		t.Fatalf("pay output=%#v saga=%#v message=%#v error=%v", output, sagas.saga, outboxWriter.message, errorValue)
	}

	result := events.PaymentResult{OrderID: order.ID, PaymentID: "payment-1", AmountCents: order.TotalAmountCents}
	envelope, _ := events.NewEnvelope("result-1", events.PaymentApprovedV1, output.SagaID, output.CorrelationID, now, result)
	handle := NewHandlePaymentResultUseCase(orders, sagas, &inboxWriterStub{created: true}, orders, orders, func() time.Time { return now.Add(time.Second) })
	if errorValue := handle.Execute(context.Background(), HandlePaymentResultInput{Envelope: envelope, Result: result, Approved: true}); errorValue != nil || order.Status != domain.OrderStatusPaid || sagas.saga.Status != domain.SagaStatusCompleted {
		t.Fatalf("handle order=%#v saga=%#v error=%v", order, sagas.saga, errorValue)
	}
}
