package usecase

import (
	"context"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
)

type CustomerReader interface {
	FindByID(context context.Context, customerID string) (*authenticationdomain.User, error)
}

type ProductReader interface {
	FindByIDs(
		context context.Context,
		productIDs []string,
	) ([]*catalogdomain.Product, error)
}

type InventoryManager interface {
	Reserve(
		context context.Context,
		stockItems []inventoryusecase.StockItem,
	) error
	Release(
		context context.Context,
		stockItems []inventoryusecase.StockItem,
	) error
}

type OrderWriter interface {
	Save(context context.Context, order *orderdomain.Order) error
	UpdateStatus(context context.Context, order *orderdomain.Order) error
}

type TransactionManager interface {
	Execute(
		context context.Context,
		operation func(transactionContext context.Context) error,
	) error
}
