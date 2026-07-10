package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyOrderID     = errors.New("the order ID must not be empty.")
	ErrEmptyOrderUserID = errors.New("the order user ID must not be empty.")
	ErrEmptyOrderItems  = errors.New("the order must contain at least one item.")
	ErrInvalidOrderItem = errors.New("the order item is invalid.")
	ErrOrderNotFound    = errors.New("the order was not found.")
)

type OrderItem struct {
	ProductID      string
	ProductName    string
	UnitPriceCents int64
	Quantity       int
}

func (orderItem OrderItem) SubtotalCents() int64 {
	return orderItem.UnitPriceCents * int64(orderItem.Quantity)
}

type Order struct {
	ID               string
	UserID           string
	Items            []OrderItem
	TotalAmountCents int64
	Status           OrderStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewOrder(
	orderID string,
	userID string,
	orderItems []OrderItem,
	createdAt time.Time,
) (*Order, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, ErrEmptyOrderID
	}

	if strings.TrimSpace(userID) == "" {
		return nil, ErrEmptyOrderUserID
	}

	if len(orderItems) == 0 {
		return nil, ErrEmptyOrderItems
	}

	var totalAmountCents int64

	for _, orderItem := range orderItems {
		if strings.TrimSpace(orderItem.ProductID) == "" ||
			strings.TrimSpace(orderItem.ProductName) == "" ||
			orderItem.UnitPriceCents <= 0 ||
			orderItem.Quantity <= 0 {
			return nil, ErrInvalidOrderItem
		}

		totalAmountCents += orderItem.SubtotalCents()
	}

	normalizedTime := createdAt.UTC()

	return &Order{
		ID:               orderID,
		UserID:           userID,
		Items:            orderItems,
		TotalAmountCents: totalAmountCents,
		Status:           OrderStatusPending,
		CreatedAt:        normalizedTime,
		UpdatedAt:        normalizedTime,
	}, nil
}

func RestoreOrder(
	orderID string,
	userID string,
	orderItems []OrderItem,
	totalAmountCents int64,
	orderStatus OrderStatus,
	createdAt time.Time,
	updatedAt time.Time,
) (*Order, error) {
	order, errorValue := NewOrder(orderID, userID, orderItems, createdAt)
	if errorValue != nil {
		return nil, errorValue
	}

	if !orderStatus.IsValid() {
		return nil, ErrInvalidOrderStatus
	}

	order.TotalAmountCents = totalAmountCents
	order.Status = orderStatus
	order.UpdatedAt = updatedAt.UTC()

	return order, nil
}

func (order *Order) MarkAsPaid(updatedAt time.Time) {
	order.Status = OrderStatusPaid
	order.UpdatedAt = updatedAt.UTC()
}

func (order *Order) MarkAsFailed(updatedAt time.Time) {
	order.Status = OrderStatusFailed
	order.UpdatedAt = updatedAt.UTC()
}
