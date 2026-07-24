package inventoryrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type StockRepository struct {
	queries *databasequeries.Queries
}

func NewStockRepository(queries *databasequeries.Queries) *StockRepository {
	return &StockRepository{queries: queries}
}

var _ usecase.StockRepository = (*StockRepository)(nil)

func (repository *StockRepository) Save(
	applicationContext context.Context,
	stock *domain.Stock,
) error {
	return transaction.Queries(
		applicationContext,
		repository.queries,
	).UpsertStock(applicationContext, databasequeries.UpsertStockParams{
		ProductID: stock.ProductID,
		Quantity:  int32(stock.Quantity),
		UpdatedAt: stock.UpdatedAt,
	})
}

func (repository *StockRepository) FindByProductID(
	applicationContext context.Context,
	productID string,
) (*domain.Stock, error) {
	stock, errorValue := transaction.Queries(
		applicationContext,
		repository.queries,
	).GetStockByProductID(applicationContext, productID)

	return restoreStock(stock, errorValue)
}

func (repository *StockRepository) FindByProductIDForUpdate(
	applicationContext context.Context,
	productID string,
) (*domain.Stock, error) {
	stock, errorValue := transaction.Queries(
		applicationContext,
		repository.queries,
	).GetStockByProductIDForUpdate(applicationContext, productID)

	return restoreStock(stock, errorValue)
}

func restoreStock(
	stock databasequeries.Stock,
	errorValue error,
) (*domain.Stock, error) {
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrStockNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return domain.NewStock(
		stock.ProductID,
		int(stock.Quantity),
		stock.UpdatedAt,
	)
}
