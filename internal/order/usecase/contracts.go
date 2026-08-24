package usecase

import (
	"context"
	"time"

	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
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

type SagaRepository interface {
	Save(context.Context, *domain.Saga) error
	FindByIDForUpdate(context.Context, string) (*domain.Saga, error)
	Update(context.Context, *domain.Saga) error
}

type OutboxWriter interface {
	Save(context.Context, *outbox.Message) error
}
type InboxWriter interface {
	TrySave(context.Context, string, time.Time) (bool, error)
}

type InventoryManager interface {
	Release(context.Context, []inventoryusecase.StockItem) error
}

type TransactionManager interface {
	Execute(context.Context, func(context.Context) error) error
}
