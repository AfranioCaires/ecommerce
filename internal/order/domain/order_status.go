package domain

import "errors"

var (
	ErrInvalidOrderStatus     = errors.New("the order status is invalid.")
	ErrInvalidOrderTransition = errors.New("the order status transition is invalid.")
)

type OrderStatus string

const (
	OrderStatusPending        OrderStatus = "PENDING"
	OrderStatusPaymentPending OrderStatus = "PAYMENT_PENDING"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCanceled       OrderStatus = "CANCELED"
)

func (orderStatus OrderStatus) IsValid() bool {
	switch orderStatus {
	case OrderStatusPending, OrderStatusPaymentPending, OrderStatusPaid, OrderStatusCanceled:
		return true
	default:
		return false
	}
}
