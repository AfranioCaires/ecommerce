package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

var errPaymentDependency = errors.New("payment dependency failed")

type paymentRepositoryFake struct {
	saveError error
	saved     *domain.Payment
}

type paymentMessageDependencies struct{ message *outbox.Message }

func (dependencies *paymentMessageDependencies) TrySave(context.Context, string, time.Time) (bool, error) {
	return true, nil
}
func (dependencies *paymentMessageDependencies) Save(_ context.Context, message *outbox.Message) error {
	dependencies.message = message
	return nil
}
func (dependencies *paymentMessageDependencies) Execute(applicationContext context.Context, operation func(context.Context) error) error {
	return operation(applicationContext)
}

func (repository *paymentRepositoryFake) Save(_ context.Context, payment *domain.Payment) error {
	if repository.saveError != nil {
		return repository.saveError
	}
	repository.saved = payment
	return nil
}

func TestPaymentServiceProcessesRequestedMessage(t *testing.T) {
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	repository := &paymentRepositoryFake{}
	dependencies := &paymentMessageDependencies{}
	service := NewMessagePaymentService(repository, &paymentGatewayFake{status: domain.PaymentStatusApproved}, dependencies, dependencies, dependencies, func() time.Time { return now })
	request := events.PaymentRequested{OrderID: "order-1", AmountCents: 1000}
	envelope, _ := events.NewEnvelope("message-1", events.PaymentRequestedV1, "saga-1", "correlation-1", now, request)
	if errorValue := service.ProcessRequested(context.Background(), envelope, request); errorValue != nil || repository.saved == nil || dependencies.message == nil || dependencies.message.MessageType != events.PaymentApprovedV1 {
		t.Fatalf("saved=%#v message=%#v error=%v", repository.saved, dependencies.message, errorValue)
	}
}

func (repository *paymentRepositoryFake) FindByOrderID(context.Context, string) (*domain.Payment, error) {
	return nil, domain.ErrPaymentNotFound
}

type paymentGatewayFake struct {
	status domain.PaymentStatus
	error  error
}

func (gateway *paymentGatewayFake) Authorize(_ context.Context, _ string, _ int64) (domain.PaymentStatus, error) {
	return gateway.status, gateway.error
}

func TestPaymentServiceProcess(t *testing.T) {
	now := time.Date(2026, time.July, 15, 15, 0, 0, 0, time.FixedZone("test", -3*60*60))
	testCases := []struct {
		name       string
		orderID    string
		amount     int64
		status     domain.PaymentStatus
		gatewayErr error
		saveErr    error
		wantErr    error
	}{
		{"gateway error", "order-1", 100, domain.PaymentStatusApproved, errPaymentDependency, nil, errPaymentDependency},
		{"invalid status", "order-1", 100, domain.PaymentStatus("UNKNOWN"), nil, nil, domain.ErrInvalidPaymentStatus},
		{"invalid amount", "order-1", 0, domain.PaymentStatusApproved, nil, nil, domain.ErrInvalidPaymentAmount},
		{"save error", "order-1", 100, domain.PaymentStatusApproved, nil, errPaymentDependency, errPaymentDependency},
		{"approved", "order-1", 100, domain.PaymentStatusApproved, nil, nil, nil},
		{"declined", "order-1", 113, domain.PaymentStatusDeclined, nil, nil, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &paymentRepositoryFake{saveError: testCase.saveErr}
			gateway := &paymentGatewayFake{status: testCase.status, error: testCase.gatewayErr}
			service := NewPaymentService(repository, gateway, func() time.Time { return now })
			payment, errorValue := service.Process(context.Background(), testCase.orderID, testCase.amount)
			if !errors.Is(errorValue, testCase.wantErr) {
				t.Fatalf("Process() error = %v, want %v", errorValue, testCase.wantErr)
			}
			if testCase.wantErr != nil {
				if payment != nil || repository.saved != nil {
					t.Fatalf("Process() = %#v; saved = %#v", payment, repository.saved)
				}
				return
			}
			if payment == nil || repository.saved != payment || payment.ID == "" || payment.OrderID != testCase.orderID || payment.AmountCents != testCase.amount || payment.Status != testCase.status || !payment.CreatedAt.Equal(now) || payment.CreatedAt.Location() != time.UTC {
				t.Fatalf("Process() = %#v; saved = %#v", payment, repository.saved)
			}
		})
	}
}
