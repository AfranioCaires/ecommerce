package orderrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/order/domain"
	"github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type OrderRepository struct {
	databaseConnection *gorm.DB
}

func NewOrderRepository(databaseConnection *gorm.DB) *OrderRepository {
	return &OrderRepository{databaseConnection: databaseConnection}
}

var _ usecase.OrderRepository = (*OrderRepository)(nil)

func (repository *OrderRepository) Save(
	context context.Context,
	order *domain.Order,
) error {
	orderModel := toOrderModel(order)
	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	return databaseConnection.Create(&orderModel).Error
}

func (repository *OrderRepository) UpdateStatus(
	context context.Context,
	order *domain.Order,
) error {
	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	return databaseConnection.
		Model(&OrderModel{}).
		Where("id = ?", order.ID).
		Updates(map[string]any{
			"status":     order.Status,
			"updated_at": order.UpdatedAt,
		}).
		Error
}

func (repository *OrderRepository) FindByID(
	context context.Context,
	orderID string,
) (*domain.Order, error) {
	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	var orderModel OrderModel

	errorValue := databaseConnection.
		Preload("Items").
		First(&orderModel, "id = ?", orderID).
		Error

	if errors.Is(errorValue, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toOrderEntity(&orderModel)
}

func (repository *OrderRepository) FindByUserID(
	context context.Context,
	userID string,
) ([]*domain.Order, error) {
	return repository.findMany(context, "user_id = ?", userID)
}

func (repository *OrderRepository) FindAll(
	context context.Context,
) ([]*domain.Order, error) {
	return repository.findMany(context, "1 = 1")
}

func (repository *OrderRepository) findMany(
	context context.Context,
	condition string,
	arguments ...any,
) ([]*domain.Order, error) {
	databaseConnection := transaction.DatabaseConnection(
		context,
		repository.databaseConnection,
	)

	var orderModels []OrderModel

	errorValue := databaseConnection.
		Preload("Items").
		Where(condition, arguments...).
		Order("created_at DESC").
		Find(&orderModels).
		Error
	if errorValue != nil {
		return nil, errorValue
	}

	orders := make([]*domain.Order, 0, len(orderModels))

	for index := range orderModels {
		order, errorValue := toOrderEntity(&orderModels[index])
		if errorValue != nil {
			return nil, errorValue
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func toOrderModel(order *domain.Order) OrderModel {
	orderItemModels := make([]OrderItemModel, 0, len(order.Items))

	for _, orderItem := range order.Items {
		orderItemModels = append(orderItemModels, OrderItemModel{
			OrderID:        order.ID,
			ProductID:      orderItem.ProductID,
			ProductName:    orderItem.ProductName,
			UnitPriceCents: orderItem.UnitPriceCents,
			Quantity:       orderItem.Quantity,
		})
	}

	return OrderModel{
		ID:               order.ID,
		UserID:           order.UserID,
		TotalAmountCents: order.TotalAmountCents,
		Status:           string(order.Status),
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
		Items:            orderItemModels,
	}
}

func toOrderEntity(orderModel *OrderModel) (*domain.Order, error) {
	orderItems := make([]domain.OrderItem, 0, len(orderModel.Items))

	for _, orderItemModel := range orderModel.Items {
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:      orderItemModel.ProductID,
			ProductName:    orderItemModel.ProductName,
			UnitPriceCents: orderItemModel.UnitPriceCents,
			Quantity:       orderItemModel.Quantity,
		})
	}

	return domain.RestoreOrder(
		orderModel.ID,
		orderModel.UserID,
		orderItems,
		orderModel.TotalAmountCents,
		domain.OrderStatus(orderModel.Status),
		orderModel.CreatedAt,
		orderModel.UpdatedAt,
	)
}
