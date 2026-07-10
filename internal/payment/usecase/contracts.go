package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
)

type PaymentRepository interface {
	Save(context context.Context, payment *domain.Payment) error
}

type PaymentGateway interface {
	Authorize(
		context context.Context,
		orderID string,
		amountCents int64,
	) (domain.PaymentStatus, error)
}
