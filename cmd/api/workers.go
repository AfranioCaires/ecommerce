package main

import (
	"context"
	"errors"

	ordermessaging "github.com/afraniocaires/ecommerce/internal/order/adapter/messaging"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

func runWorkers(applicationContext context.Context, publisher *outbox.Publisher, consumer *platformmessaging.Consumer, handler *ordermessaging.PaymentResultHandler) error {
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- publisher.Run(applicationContext) }()
	go func() { errorsChannel <- consumer.Run(applicationContext, handler.Handle) }()
	errorValue := <-errorsChannel
	if errors.Is(errorValue, context.Canceled) {
		return nil
	}
	return errorValue
}
