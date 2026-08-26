package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type PaymentService struct {
	paymentRepository PaymentRepository
	paymentGateway    PaymentGateway
	currentTime       func() time.Time
	inbox             InboxWriter
	outbox            OutboxWriter
	transaction       TransactionManager
}

func NewMessagePaymentService(paymentRepository PaymentRepository, paymentGateway PaymentGateway, inbox InboxWriter, outboxWriter OutboxWriter, transaction TransactionManager, currentTime func() time.Time) *PaymentService {
	return &PaymentService{paymentRepository: paymentRepository, paymentGateway: paymentGateway, inbox: inbox, outbox: outboxWriter, transaction: transaction, currentTime: currentTime}
}

func NewPaymentService(
	paymentRepository PaymentRepository,
	paymentGateway PaymentGateway,
	currentTime func() time.Time,
) *PaymentService {
	return &PaymentService{
		paymentRepository: paymentRepository,
		paymentGateway:    paymentGateway,
		currentTime:       currentTime,
	}
}

func (service *PaymentService) ProcessRequested(applicationContext context.Context, envelope events.Envelope, request events.PaymentRequested) error {
	if envelope.MessageType != events.PaymentRequestedV1 || request.Validate() != nil {
		return events.ErrInvalidPaymentMessage
	}
	return service.transaction.Execute(applicationContext, func(transactionContext context.Context) error {
		now := service.currentTime().UTC()
		created, errorValue := service.inbox.TrySave(transactionContext, envelope.MessageID, now)
		if errorValue != nil || !created {
			return errorValue
		}
		if _, errorValue := service.paymentRepository.FindByOrderID(transactionContext, request.OrderID); errorValue == nil {
			return nil
		} else if !errors.Is(errorValue, domain.ErrPaymentNotFound) {
			return errorValue
		}
		status, errorValue := service.paymentGateway.Authorize(transactionContext, request.OrderID, request.AmountCents)
		if errorValue != nil {
			return errorValue
		}
		payment, errorValue := domain.NewPayment(uuid.NewString(), request.OrderID, request.AmountCents, status, now)
		if errorValue != nil {
			return errorValue
		}
		if errorValue := service.paymentRepository.Save(transactionContext, payment); errorValue != nil {
			return errorValue
		}
		messageType, routingKey, reasonCode := events.PaymentApprovedV1, "payment.approved", ""
		if status == domain.PaymentStatusDeclined {
			messageType, routingKey, reasonCode = events.PaymentDeclinedV1, "payment.declined", "SIMULATED_DECLINE"
		}
		result := events.PaymentResult{OrderID: payment.OrderID, PaymentID: payment.ID, AmountCents: payment.AmountCents, ReasonCode: reasonCode}
		resultEnvelope, errorValue := events.NewEnvelope(uuid.NewString(), messageType, envelope.SagaID, envelope.CorrelationID, now, result)
		if errorValue != nil {
			return errorValue
		}
		body, errorValue := json.Marshal(resultEnvelope)
		if errorValue != nil {
			return errorValue
		}
		return service.outbox.Save(transactionContext, outbox.NewMessage(resultEnvelope.MessageID, messageType, routingKey, body, now))
	})
}

func (service *PaymentService) Process(
	context context.Context,
	orderID string,
	amountCents int64,
) (*domain.Payment, error) {
	paymentStatus, errorValue := service.paymentGateway.Authorize(
		context,
		orderID,
		amountCents,
	)
	if errorValue != nil {
		return nil, errorValue
	}

	payment, errorValue := domain.NewPayment(
		uuid.NewString(),
		orderID,
		amountCents,
		paymentStatus,
		service.currentTime(),
	)
	if errorValue != nil {
		return nil, errorValue
	}

	if errorValue := service.paymentRepository.Save(
		context,
		payment,
	); errorValue != nil {
		return nil, errorValue
	}

	return payment, nil
}
