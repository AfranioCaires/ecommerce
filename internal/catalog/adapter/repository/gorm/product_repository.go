package catalogrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type ProductRepository struct {
	databaseConnection *gorm.DB
}

func NewProductRepository(databaseConnection *gorm.DB) *ProductRepository {
	return &ProductRepository{databaseConnection: databaseConnection}
}

var _ usecase.ProductRepository = (*ProductRepository)(nil)

func (repository *ProductRepository) Save(
	context context.Context,
	product *domain.Product,
) error {
	productModel := toProductModel(product)

	return repository.databaseConnection.
		WithContext(context).
		Save(&productModel).
		Error
}

func (repository *ProductRepository) FindByID(
	context context.Context,
	productID string,
) (*domain.Product, error) {
	var productModel ProductModel

	errorValue := repository.databaseConnection.
		WithContext(context).
		First(&productModel, "id = ?", productID).
		Error

	if errors.Is(errorValue, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toProductEntity(&productModel)
}

func (repository *ProductRepository) FindByIDs(
	context context.Context,
	productIDs []string,
) ([]*domain.Product, error) {
	var productModels []ProductModel

	errorValue := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	).
		Where("id IN ?", productIDs).
		Find(&productModels).
		Error
	if errorValue != nil {
		return nil, errorValue
	}

	products := make([]*domain.Product, 0, len(productModels))

	for index := range productModels {
		product, errorValue := toProductEntity(&productModels[index])
		if errorValue != nil {
			return nil, errorValue
		}

		products = append(products, product)
	}

	return products, nil
}

func (repository *ProductRepository) FindPage(
	context context.Context,
	pageRequest usecase.ProductPageRequest,
) ([]*domain.Product, int64, error) {
	var totalItems int64

	countQuery := repository.databaseConnection.
		WithContext(context).
		Model(&ProductModel{}).
		Where("status = ?", string(domain.ProductStatusActive)).
		Count(&totalItems)
	if countQuery.Error != nil {
		return nil, 0, countQuery.Error
	}

	offset := (pageRequest.PageNumber - 1) * pageRequest.PageSize

	var productModels []ProductModel

	findQuery := repository.databaseConnection.
		WithContext(context).
		Where("status = ?", string(domain.ProductStatusActive)).
		Order("created_at DESC").
		Limit(pageRequest.PageSize).
		Offset(offset).
		Find(&productModels)
	if findQuery.Error != nil {
		return nil, 0, findQuery.Error
	}

	products := make([]*domain.Product, 0, len(productModels))

	for index := range productModels {
		product, errorValue := toProductEntity(&productModels[index])
		if errorValue != nil {
			return nil, 0, errorValue
		}

		products = append(products, product)
	}

	return products, totalItems, nil
}
