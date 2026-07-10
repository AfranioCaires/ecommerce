package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
)

type SimulatedPaymentGateway struct{}

func NewSimulatedPaymentGateway() *SimulatedPaymentGateway {
	return &SimulatedPaymentGateway{}
}

func (paymentGateway *SimulatedPaymentGateway) Authorize(
	context context.Context,
	orderID string,
	amountCents int64,
) (domain.PaymentStatus, error) {
	if amountCents%100 == 13 {
		return domain.PaymentStatusDeclined, nil
	}

	return domain.PaymentStatusApproved, nil
}
