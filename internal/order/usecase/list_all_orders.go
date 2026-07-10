package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type ListAllOrdersUseCase struct {
	orderRepository OrderRepository
}

func NewListAllOrdersUseCase(
	orderRepository OrderRepository,
) *ListAllOrdersUseCase {
	return &ListAllOrdersUseCase{orderRepository: orderRepository}
}

func (useCase *ListAllOrdersUseCase) Execute(
	context context.Context,
) ([]*domain.Order, error) {
	return useCase.orderRepository.FindAll(context)
}
