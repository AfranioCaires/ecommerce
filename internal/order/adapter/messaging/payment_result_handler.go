package ordermessaging

import (
	"context"
	"errors"

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
		return platformmessaging.Permanent(errorValue)
	}
	approved := envelope.MessageType == events.PaymentApprovedV1
	if !approved && envelope.MessageType != events.PaymentDeclinedV1 {
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	result, errorValue := events.PayloadAs[events.PaymentResult](envelope)
	if errorValue != nil || result.Validate() != nil {
		return platformmessaging.Permanent(events.ErrInvalidPaymentMessage)
	}
	errorValue = handler.useCase.Execute(applicationContext, orderusecase.HandlePaymentResultInput{Envelope: envelope, Result: result, Approved: approved})
	if errors.Is(errorValue, orderusecase.ErrPaymentResultMismatch) {
		return platformmessaging.Permanent(errorValue)
	}
	return errorValue
}
