package orderrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type SagaRepository struct{ queries *databasequeries.Queries }

func NewSagaRepository(queries *databasequeries.Queries) *SagaRepository {
	return &SagaRepository{queries: queries}
}

func (repository *SagaRepository) Save(applicationContext context.Context, saga *domain.Saga) error {
	return transaction.Queries(applicationContext, repository.queries).CreateOrderSaga(applicationContext, databasequeries.CreateOrderSagaParams{
		ID: saga.ID, OrderID: saga.OrderID, CorrelationID: saga.CorrelationID, Status: string(saga.Status), CreatedAt: saga.CreatedAt, UpdatedAt: saga.UpdatedAt,
	})
}

func (repository *SagaRepository) FindByIDForUpdate(applicationContext context.Context, sagaID string) (*domain.Saga, error) {
	row, errorValue := transaction.Queries(applicationContext, repository.queries).GetOrderSagaForUpdate(applicationContext, sagaID)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrInvalidSaga
	}
	if errorValue != nil {
		return nil, errorValue
	}
	return domain.RestoreSaga(row.ID, row.OrderID, row.CorrelationID, domain.SagaStatus(row.Status), row.CreatedAt, row.UpdatedAt)
}

func (repository *SagaRepository) Update(applicationContext context.Context, saga *domain.Saga) error {
	return transaction.Queries(applicationContext, repository.queries).UpdateOrderSaga(applicationContext, databasequeries.UpdateOrderSagaParams{ID: saga.ID, Status: string(saga.Status), UpdatedAt: saga.UpdatedAt})
}
