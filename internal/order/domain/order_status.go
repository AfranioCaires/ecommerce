package domain

import "errors"

var ErrInvalidOrderStatus = errors.New("the order status is invalid.")

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "PENDING"
	OrderStatusPaid    OrderStatus = "PAID"
	OrderStatusFailed  OrderStatus = "FAILED"
)

func (orderStatus OrderStatus) IsValid() bool {
	switch orderStatus {
	case OrderStatusPending, OrderStatusPaid, OrderStatusFailed:
		return true
	default:
		return false
	}
}
