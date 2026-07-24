package paymentrepository

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	"github.com/afraniocaires/ecommerce/internal/payment/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type PaymentRepository struct {
	queries *databasequeries.Queries
}

func NewPaymentRepository(queries *databasequeries.Queries) *PaymentRepository {
	return &PaymentRepository{queries: queries}
}

var _ usecase.PaymentRepository = (*PaymentRepository)(nil)

func (repository *PaymentRepository) Save(
	applicationContext context.Context,
	payment *domain.Payment,
) error {
	return transaction.Queries(
		applicationContext,
		repository.queries,
	).CreatePayment(applicationContext, databasequeries.CreatePaymentParams{
		ID:          payment.ID,
		OrderID:     payment.OrderID,
		AmountCents: payment.AmountCents,
		Status:      string(payment.Status),
		CreatedAt:   payment.CreatedAt,
	})
}
