package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
)

type StockRepository interface {
	Save(context context.Context, stock *domain.Stock) error
	FindByProductID(
		context context.Context,
		productID string,
	) (*domain.Stock, error)
	FindByProductIDForUpdate(
		context context.Context,
		productID string,
	) (*domain.Stock, error)
}
