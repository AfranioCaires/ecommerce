package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
)

type ProductRepository interface {
	Save(context context.Context, product *domain.Product) error
	FindByID(context context.Context, productID string) (*domain.Product, error)
	FindByIDs(
		context context.Context,
		productIDs []string,
	) ([]*domain.Product, error)
	FindPage(
		context context.Context,
		pageRequest ProductPageRequest,
	) ([]*domain.Product, int64, error)
}
