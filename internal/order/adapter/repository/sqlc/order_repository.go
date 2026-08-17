package orderrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/order/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type OrderRepository struct {
	queries *databasequeries.Queries
}

func NewOrderRepository(queries *databasequeries.Queries) *OrderRepository {
	return &OrderRepository{queries: queries}
}

var _ usecase.OrderRepository = (*OrderRepository)(nil)

func (repository *OrderRepository) Save(
	applicationContext context.Context,
	order *domain.Order,
) error {
	queries := transaction.Queries(applicationContext, repository.queries)

	if errorValue := queries.CreateOrder(
		applicationContext,
		databasequeries.CreateOrderParams{
			ID:               order.ID,
			CustomerID:       order.UserID,
			TotalAmountCents: order.TotalAmountCents,
			Status:           string(order.Status),
			CreatedAt:        order.CreatedAt,
			UpdatedAt:        order.UpdatedAt,
		},
	); errorValue != nil {
		return errorValue
	}

	for _, orderItem := range order.Items {
		if errorValue := queries.CreateOrderItem(
			applicationContext,
			databasequeries.CreateOrderItemParams{
				OrderID:        order.ID,
				ProductID:      orderItem.ProductID,
				ProductName:    orderItem.ProductName,
				UnitPriceCents: orderItem.UnitPriceCents,
				Quantity:       int32(orderItem.Quantity),
			},
		); errorValue != nil {
			return errorValue
		}
	}

	return nil
}

func (repository *OrderRepository) UpdateStatus(
	applicationContext context.Context,
	order *domain.Order,
) error {
	return transaction.Queries(
		applicationContext,
		repository.queries,
	).UpdateOrderStatus(applicationContext, databasequeries.UpdateOrderStatusParams{
		ID:        order.ID,
		Status:    string(order.Status),
		UpdatedAt: order.UpdatedAt,
	})
}

func (repository *OrderRepository) FindByID(
	applicationContext context.Context,
	orderID string,
) (*domain.Order, error) {
	queries := transaction.Queries(applicationContext, repository.queries)

	order, errorValue := queries.GetOrderByID(applicationContext, orderID)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	items, errorValue := queries.ListOrderItemsByOrderIDs(
		applicationContext,
		[]string{order.ID},
	)
	if errorValue != nil {
		return nil, errorValue
	}

	return restoreOrder(order, items)
}

func (repository *OrderRepository) FindByIDForUpdate(
	applicationContext context.Context,
	orderID string,
) (*domain.Order, error) {
	queries := transaction.Queries(applicationContext, repository.queries)
	order, errorValue := queries.GetOrderByIDForUpdate(applicationContext, orderID)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}
	items, errorValue := queries.ListOrderItemsByOrderIDs(applicationContext, []string{order.ID})
	if errorValue != nil {
		return nil, errorValue
	}
	return restoreOrder(order, items)
}

func (repository *OrderRepository) FindByUserID(
	applicationContext context.Context,
	userID string,
	pageRequest usecase.OrderPageRequest,
) ([]*domain.Order, error) {
	queries := transaction.Queries(applicationContext, repository.queries)

	orders, errorValue := queries.ListOrdersByCustomer(
		applicationContext,
		databasequeries.ListOrdersByCustomerParams{
			CustomerID: userID,
			PageLimit:  int32(pageRequest.Limit),
			PageOffset: int32(pageRequest.Offset),
		},
	)
	if errorValue != nil {
		return nil, errorValue
	}

	return repository.restoreOrders(applicationContext, queries, orders)
}

func (repository *OrderRepository) FindAll(
	applicationContext context.Context,
	pageRequest usecase.OrderPageRequest,
) ([]*domain.Order, error) {
	queries := transaction.Queries(applicationContext, repository.queries)

	orders, errorValue := queries.ListOrders(
		applicationContext,
		databasequeries.ListOrdersParams{
			PageLimit:  int32(pageRequest.Limit),
			PageOffset: int32(pageRequest.Offset),
		},
	)
	if errorValue != nil {
		return nil, errorValue
	}

	return repository.restoreOrders(applicationContext, queries, orders)
}

func (repository *OrderRepository) restoreOrders(
	applicationContext context.Context,
	queries *databasequeries.Queries,
	orders []databasequeries.Order,
) ([]*domain.Order, error) {
	if len(orders) == 0 {
		return []*domain.Order{}, nil
	}

	orderIDs := make([]string, len(orders))
	for index, order := range orders {
		orderIDs[index] = order.ID
	}

	items, errorValue := queries.ListOrderItemsByOrderIDs(
		applicationContext,
		orderIDs,
	)
	if errorValue != nil {
		return nil, errorValue
	}

	itemsByOrderID := make(map[string][]databasequeries.OrderItem, len(orders))
	for _, item := range items {
		itemsByOrderID[item.OrderID] = append(itemsByOrderID[item.OrderID], item)
	}

	orderEntities := make([]*domain.Order, 0, len(orders))
	for _, order := range orders {
		orderEntity, errorValue := restoreOrder(order, itemsByOrderID[order.ID])
		if errorValue != nil {
			return nil, errorValue
		}
		orderEntities = append(orderEntities, orderEntity)
	}

	return orderEntities, nil
}

func restoreOrder(
	order databasequeries.Order,
	items []databasequeries.OrderItem,
) (*domain.Order, error) {
	orderItems := make([]domain.OrderItem, 0, len(items))
	for _, item := range items {
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:      item.ProductID,
			ProductName:    item.ProductName,
			UnitPriceCents: item.UnitPriceCents,
			Quantity:       int(item.Quantity),
		})
	}

	return domain.RestoreOrder(
		order.ID,
		order.CustomerID,
		orderItems,
		order.TotalAmountCents,
		domain.OrderStatus(order.Status),
		order.CreatedAt,
		order.UpdatedAt,
	)
}
