package catalogrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type ProductRepository struct {
	queries *databasequeries.Queries
}

func NewProductRepository(queries *databasequeries.Queries) *ProductRepository {
	return &ProductRepository{queries: queries}
}

var _ usecase.ProductRepository = (*ProductRepository)(nil)

func (repository *ProductRepository) Save(
	applicationContext context.Context,
	product *domain.Product,
) error {
	return repository.queries.UpsertProduct(
		applicationContext,
		databasequeries.UpsertProductParams{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			PriceCents:  product.PriceCents,
			Status:      string(product.Status),
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		},
	)
}

func (repository *ProductRepository) FindByID(
	applicationContext context.Context,
	productID string,
) (*domain.Product, error) {
	product, errorValue := repository.queries.GetProductByID(
		applicationContext,
		productID,
	)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toProduct(product)
}

func (repository *ProductRepository) FindByIDs(
	applicationContext context.Context,
	productIDs []string,
) ([]*domain.Product, error) {
	products, errorValue := transaction.Queries(
		applicationContext,
		repository.queries,
	).ListProductsByIDs(applicationContext, productIDs)
	if errorValue != nil {
		return nil, errorValue
	}

	return toProducts(products)
}

func (repository *ProductRepository) FindPage(
	applicationContext context.Context,
	pageRequest usecase.ProductPageRequest,
) ([]*domain.Product, int64, error) {
	totalItems, errorValue := repository.queries.CountActiveProducts(
		applicationContext,
	)
	if errorValue != nil {
		return nil, 0, errorValue
	}

	products, errorValue := repository.queries.ListActiveProducts(
		applicationContext,
		databasequeries.ListActiveProductsParams{
			PageLimit:  int32(pageRequest.PageSize),
			PageOffset: int32((pageRequest.PageNumber - 1) * pageRequest.PageSize),
		},
	)
	if errorValue != nil {
		return nil, 0, errorValue
	}

	productEntities, errorValue := toProducts(products)
	if errorValue != nil {
		return nil, 0, errorValue
	}

	return productEntities, totalItems, nil
}

func toProducts(products []databasequeries.Product) ([]*domain.Product, error) {
	productEntities := make([]*domain.Product, 0, len(products))
	for _, product := range products {
		productEntity, errorValue := toProduct(product)
		if errorValue != nil {
			return nil, errorValue
		}
		productEntities = append(productEntities, productEntity)
	}

	return productEntities, nil
}

func toProduct(product databasequeries.Product) (*domain.Product, error) {
	return domain.RestoreProduct(
		product.ID,
		product.Name,
		product.Description,
		product.PriceCents,
		domain.ProductStatus(product.Status),
		product.CreatedAt,
		product.UpdatedAt,
	)
}
