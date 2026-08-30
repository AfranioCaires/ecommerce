package ordermessaging

import (
	"context"
	"errors"
	"log/slog"

	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/events"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
)

type PaymentResultHandler struct {
	useCase *orderusecase.HandlePaymentResultUseCase
}

func NewPaymentResultHandler(useCase *orderusecase.HandlePaymentResultUseCase) *PaymentResultHandler {
	return &PaymentResultHandler{useCase: useCase}
}

func (handler *PaymentResultHandler) Handle(applicationContext context.Context, body []byte) error {
	envelope, errorValue := events.Decode(body)
	if errorValue != nil {
		slog.Warn("Payment result rejected.", "operation", "payment-result.consume", "result", "invalid", "error", errorValue)
		return platformmessaging.Permanent(errorValue)
	}
	approved := envelope.MessageType == events.PaymentApprovedV1
	if !approved && envelope.MessageType != events.PaymentDeclinedV1 {
		slog.Warn("Payment result rejected.", "operation", "payment-result.consume", "result", "invalid", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID)
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	result, errorValue := events.PayloadAs[events.PaymentResult](envelope)
	if errorValue != nil || result.Validate() != nil {
		slog.Warn("Payment result rejected.", "operation", "payment-result.consume", "result", "invalid", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID)
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	errorValue = handler.useCase.Execute(applicationContext, orderusecase.HandlePaymentResultInput{Envelope: envelope, Result: result, Approved: approved})
	if errors.Is(errorValue, orderusecase.ErrPaymentResultMismatch) {
		slog.Warn("Payment result rejected.", "operation", "payment-result.consume", "result", "mismatch", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID, "order_id", result.OrderID, "payment_id", result.PaymentID)
		return platformmessaging.Permanent(errorValue)
	}
	if errorValue != nil {
		slog.Error("Payment result processing failed.", "operation", "payment-result.consume", "result", "failed", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID, "order_id", result.OrderID, "payment_id", result.PaymentID, "error", errorValue)
		return errorValue
	}
	slog.Info("Payment result processed.", "operation", "payment-result.consume", "result", "success", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID, "order_id", result.OrderID, "payment_id", result.PaymentID)
	return errorValue
}
