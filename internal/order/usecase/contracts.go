package usecase

import (
	"context"

	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type OrderRepository interface {
	Save(context context.Context, order *domain.Order) error
	UpdateStatus(context context.Context, order *domain.Order) error
	FindByID(context context.Context, orderID string) (*domain.Order, error)
	FindByIDForUpdate(context context.Context, orderID string) (*domain.Order, error)
	FindByUserID(
		context context.Context,
		userID string,
		pageRequest OrderPageRequest,
	) ([]*domain.Order, error)
	FindAll(
		context context.Context,
		pageRequest OrderPageRequest,
	) ([]*domain.Order, error)
}

type InventoryManager interface {
	Release(context.Context, []inventoryusecase.StockItem) error
}

type TransactionManager interface {
	Execute(context.Context, func(context.Context) error) error
}
