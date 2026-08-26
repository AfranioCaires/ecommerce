package usecase

import (
	"context"
	"time"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type PaymentRepository interface {
	Save(context context.Context, payment *domain.Payment) error
	FindByOrderID(context context.Context, orderID string) (*domain.Payment, error)
}

type InboxWriter interface {
	TrySave(context.Context, string, time.Time) (bool, error)
}
type OutboxWriter interface {
	Save(context.Context, *outbox.Message) error
}
type TransactionManager interface {
	Execute(context.Context, func(context.Context) error) error
}

type PaymentGateway interface {
	Authorize(
		context context.Context,
		orderID string,
		amountCents int64,
	) (domain.PaymentStatus, error)
}
