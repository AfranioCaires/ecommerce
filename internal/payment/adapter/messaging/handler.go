package paymentmessaging

import (
	"context"
	"log/slog"

	"github.com/afraniocaires/ecommerce/internal/platform/events"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
)

type PaymentRequestProcessor interface {
	ProcessRequested(context.Context, events.Envelope, events.PaymentRequested) error
}

type PaymentRequestHandler struct {
	processor PaymentRequestProcessor
	logger    *slog.Logger
}

func NewPaymentRequestHandler(processor PaymentRequestProcessor, logger *slog.Logger) *PaymentRequestHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PaymentRequestHandler{processor: processor, logger: logger}
}

func (handler *PaymentRequestHandler) Handle(applicationContext context.Context, body []byte) error {
	envelope, errorValue := events.Decode(body)
	if errorValue != nil {
		handler.logger.Warn("Payment request rejected.", "operation", "payment.consume", "result", "invalid", "error", errorValue)
		return platformmessaging.Permanent(errorValue)
	}
	if envelope.MessageType != events.PaymentRequestedV1 {
		handler.logger.Warn("Payment request rejected.", "operation", "payment.consume", "result", "invalid", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID)
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	request, errorValue := events.PayloadAs[events.PaymentRequested](envelope)
	if errorValue != nil || request.Validate() != nil {
		handler.logger.Warn("Payment request rejected.", "operation", "payment.consume", "result", "invalid", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID)
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	if errorValue := handler.processor.ProcessRequested(applicationContext, envelope, request); errorValue != nil {
		handler.logger.Error("Payment request processing failed.", "operation", "payment.consume", "result", "failed", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID, "order_id", request.OrderID, "error", errorValue)
		return errorValue
	}
	handler.logger.Info("Payment request processed.", "operation", "payment.consume", "result", "success", "message_id", envelope.MessageID, "message_type", envelope.MessageType, "saga_id", envelope.SagaID, "correlation_id", envelope.CorrelationID, "order_id", request.OrderID)
	return nil
}
