package paymentrepository

import (
	"context"

	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/payment/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type PaymentRepository struct {
	databaseConnection *gorm.DB
}

func NewPaymentRepository(databaseConnection *gorm.DB) *PaymentRepository {
	return &PaymentRepository{databaseConnection: databaseConnection}
}

var _ usecase.PaymentRepository = (*PaymentRepository)(nil)

func (repository *PaymentRepository) Save(
	context context.Context,
	payment *domain.Payment,
) error {
	paymentModel := PaymentModel{
		ID:          payment.ID,
		OrderID:     payment.OrderID,
		AmountCents: payment.AmountCents,
		Status:      string(payment.Status),
		CreatedAt:   payment.CreatedAt,
	}

	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	return databaseConnection.Create(&paymentModel).Error
}
