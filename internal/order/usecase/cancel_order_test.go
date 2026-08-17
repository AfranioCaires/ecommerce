package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type cancelDependencies struct {
	order    *domain.Order
	released []inventoryusecase.StockItem
	updated  bool
}

func (dependencies *cancelDependencies) Save(context.Context, *domain.Order) error { return nil }
func (dependencies *cancelDependencies) UpdateStatus(_ context.Context, order *domain.Order) error {
	dependencies.order, dependencies.updated = order, true
	return nil
}
func (dependencies *cancelDependencies) FindByID(context.Context, string) (*domain.Order, error) {
	return dependencies.order, nil
}
func (dependencies *cancelDependencies) FindByIDForUpdate(context.Context, string) (*domain.Order, error) {
	if dependencies.order == nil {
		return nil, domain.ErrOrderNotFound
	}
	return dependencies.order, nil
}
func (dependencies *cancelDependencies) FindByUserID(context.Context, string, OrderPageRequest) ([]*domain.Order, error) {
	return nil, nil
}
func (dependencies *cancelDependencies) FindAll(context.Context, OrderPageRequest) ([]*domain.Order, error) {
	return nil, nil
}
func (dependencies *cancelDependencies) Release(_ context.Context, items []inventoryusecase.StockItem) error {
	dependencies.released = append([]inventoryusecase.StockItem(nil), items...)
	return nil
}
func (dependencies *cancelDependencies) Execute(applicationContext context.Context, operation func(context.Context) error) error {
	return operation(applicationContext)
}

func TestCancelOrderUseCase(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	order, _ := domain.NewOrder("order-1", "customer-1", []domain.OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1000, Quantity: 2}}, now)
	dependencies := &cancelDependencies{order: order}
	useCase := NewCancelOrderUseCase(dependencies, dependencies, dependencies, func() time.Time { return now.Add(time.Hour) })

	canceled, errorValue := useCase.Execute(context.Background(), CancelOrderInput{OrderID: order.ID, UserID: order.UserID})
	if errorValue != nil || canceled.Status != domain.OrderStatusCanceled || !dependencies.updated || len(dependencies.released) != 1 || dependencies.released[0].Quantity != 2 {
		t.Fatalf("Execute() = %#v, %v, released=%#v", canceled, errorValue, dependencies.released)
	}

	if _, errorValue := useCase.Execute(context.Background(), CancelOrderInput{OrderID: order.ID, UserID: order.UserID}); !errors.Is(errorValue, domain.ErrInvalidOrderTransition) {
		t.Fatalf("expected transition conflict, received %v", errorValue)
	}
}
