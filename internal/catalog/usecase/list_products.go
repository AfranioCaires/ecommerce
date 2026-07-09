package usecase

import (
	"context"
)

type ListProductsUseCase struct {
	productRepository ProductRepository
}

func NewListProductsUseCase(
	productRepository ProductRepository,
) *ListProductsUseCase {
	return &ListProductsUseCase{
		productRepository: productRepository,
	}
}

func (useCase *ListProductsUseCase) Execute(
	context context.Context,
	pageRequest ProductPageRequest,
) (*ProductPage, error) {
	products, totalItems, errorValue := useCase.productRepository.FindPage(
		context,
		pageRequest,
	)
	if errorValue != nil {
		return nil, errorValue
	}

	productPage := NewProductPage(products, pageRequest, totalItems)
	return &productPage, nil
}
