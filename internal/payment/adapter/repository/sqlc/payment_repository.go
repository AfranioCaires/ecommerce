package paymentrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	databasequeries "github.com/afraniocaires/ecommerce/internal/payment/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/payment/platform/transaction"
	"github.com/afraniocaires/ecommerce/internal/payment/usecase"
)

type PaymentRepository struct {
	queries *databasequeries.Queries
}

func (repository *PaymentRepository) FindByOrderID(applicationContext context.Context, orderID string) (*domain.Payment, error) {
	row, errorValue := transaction.Queries(applicationContext, repository.queries).GetPaymentByOrderID(applicationContext, orderID)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}
	return domain.NewPayment(row.ID, row.OrderID, row.AmountCents, domain.PaymentStatus(row.Status), row.CreatedAt)
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
