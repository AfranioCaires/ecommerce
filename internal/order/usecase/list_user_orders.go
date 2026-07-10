package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type ListUserOrdersUseCase struct {
	orderRepository OrderRepository
}

func NewListUserOrdersUseCase(
	orderRepository OrderRepository,
) *ListUserOrdersUseCase {
	return &ListUserOrdersUseCase{orderRepository: orderRepository}
}

func (useCase *ListUserOrdersUseCase) Execute(
	context context.Context,
	userID string,
) ([]*domain.Order, error) {
	return useCase.orderRepository.FindByUserID(context, userID)
}
