package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
)

type queryRepository struct {
	order       *domain.Order
	orders      []*domain.Order
	err         error
	orderID     string
	userID      string
	pageRequest OrderPageRequest
	method      string
}

func (repository *queryRepository) Save(context.Context, *domain.Order) error         { return nil }
func (repository *queryRepository) UpdateStatus(context.Context, *domain.Order) error { return nil }
func (repository *queryRepository) FindByID(_ context.Context, orderID string) (*domain.Order, error) {
	repository.method, repository.orderID = "FindByID", orderID
	return repository.order, repository.err
}
func (repository *queryRepository) FindByUserID(_ context.Context, userID string, pageRequest OrderPageRequest) ([]*domain.Order, error) {
	repository.method, repository.userID, repository.pageRequest = "FindByUserID", userID, pageRequest
	return repository.orders, repository.err
}
func (repository *queryRepository) FindAll(_ context.Context, pageRequest OrderPageRequest) ([]*domain.Order, error) {
	repository.method, repository.pageRequest = "FindAll", pageRequest
	return repository.orders, repository.err
}

func TestOrderQueries(t *testing.T) {
	dependencyError := errors.New("dependency failed")
	pageRequest := OrderPageRequest{Limit: 12, Offset: 3}

	for _, returnedError := range []error{nil, dependencyError} {
		repository := &queryRepository{err: returnedError}
		got, err := NewGetOrderUseCase(repository).Execute(context.Background(), "order-1")
		if got != repository.order || !errors.Is(err, returnedError) || repository.method != "FindByID" || repository.orderID != "order-1" {
			t.Fatalf("GetOrder Execute() = %#v, %v; repository = %#v", got, err, repository)
		}

		repository = &queryRepository{orders: []*domain.Order{{ID: "order-1"}}, err: returnedError}
		orders, err := NewListUserOrdersUseCase(repository).Execute(context.Background(), "user-1", pageRequest)
		if len(orders) != len(repository.orders) || !errors.Is(err, returnedError) || repository.method != "FindByUserID" || repository.userID != "user-1" || repository.pageRequest != pageRequest {
			t.Fatalf("ListUserOrders Execute() = %#v, %v; repository = %#v", orders, err, repository)
		}

		repository = &queryRepository{orders: []*domain.Order{{ID: "order-1"}}, err: returnedError}
		orders, err = NewListAllOrdersUseCase(repository).Execute(context.Background(), pageRequest)
		if len(orders) != len(repository.orders) || !errors.Is(err, returnedError) || repository.method != "FindAll" || repository.pageRequest != pageRequest {
			t.Fatalf("ListAllOrders Execute() = %#v, %v; repository = %#v", orders, err, repository)
		}
	}
}
