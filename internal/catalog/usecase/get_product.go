package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
)

type GetProductUseCase struct {
	productRepository ProductRepository
}

func NewGetProductUseCase(
	productRepository ProductRepository,
) *GetProductUseCase {
	return &GetProductUseCase{productRepository: productRepository}
}

func (useCase *GetProductUseCase) Execute(
	context context.Context,
	productID string,
) (*domain.Product, error) {
	return useCase.productRepository.FindByID(context, productID)
}
