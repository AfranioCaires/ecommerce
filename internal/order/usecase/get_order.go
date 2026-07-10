package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type GetOrderUseCase struct {
	orderRepository OrderRepository
}

func NewGetOrderUseCase(
	orderRepository OrderRepository,
) *GetOrderUseCase {
	return &GetOrderUseCase{orderRepository: orderRepository}
}

func (useCase *GetOrderUseCase) Execute(
	context context.Context,
	orderID string,
) (*domain.Order, error) {
	return useCase.orderRepository.FindByID(context, orderID)
}
