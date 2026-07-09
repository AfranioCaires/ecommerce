package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
)

type CreateProductInput struct {
	Name        string
	Description string
	PriceCents  int64
}

type CreateProductUseCase struct {
	productRepository ProductRepository
	currentTime       func() time.Time
}

func NewCreateProductUseCase(
	productRepository ProductRepository,
	currentTime func() time.Time,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepository: productRepository,
		currentTime:       currentTime,
	}
}

func (useCase *CreateProductUseCase) Execute(
	context context.Context,
	input CreateProductInput,
) (*domain.Product, error) {
	product, errorValue := domain.NewProduct(
		uuid.NewString(),
		input.Name,
		input.Description,
		input.PriceCents,
		useCase.currentTime(),
	)
	if errorValue != nil {
		return nil, errorValue
	}

	if errorValue := useCase.productRepository.Save(
		context,
		product,
	); errorValue != nil {
		return nil, errorValue
	}

	return product, nil
}
