package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
)

type PaymentService struct {
	paymentRepository PaymentRepository
	paymentGateway    PaymentGateway
	currentTime       func() time.Time
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
