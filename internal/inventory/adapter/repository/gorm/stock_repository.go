package inventoryrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
	"github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type StockRepository struct {
	databaseConnection *gorm.DB
}

func NewStockRepository(databaseConnection *gorm.DB) *StockRepository {
	return &StockRepository{databaseConnection: databaseConnection}
}

var _ usecase.StockRepository = (*StockRepository)(nil)

func (repository *StockRepository) Save(
	context context.Context,
	stock *domain.Stock,
) error {
	stockModel := StockModel{
		ProductID: stock.ProductID,
		Quantity:  stock.Quantity,
		UpdatedAt: stock.UpdatedAt,
	}

	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	return databaseConnection.Save(&stockModel).Error
}

func (repository *StockRepository) FindByProductID(
	context context.Context,
	productID string,
) (*domain.Stock, error) {
	return repository.find(context, productID, false)
}

func (repository *StockRepository) FindByProductIDForUpdate(
	context context.Context,
	productID string,
) (*domain.Stock, error) {
	return repository.find(context, productID, true)
}

func (repository *StockRepository) find(
	context context.Context,
	productID string,
	lockForUpdate bool,
) (*domain.Stock, error) {
	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	query := databaseConnection
	if lockForUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var stockModel StockModel

	errorValue := query.
		First(&stockModel, "product_id = ?", productID).
		Error

	if errors.Is(errorValue, gorm.ErrRecordNotFound) {
		return nil, domain.ErrStockNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return domain.NewStock(
		stockModel.ProductID,
		stockModel.Quantity,
		stockModel.UpdatedAt,
	)
}
