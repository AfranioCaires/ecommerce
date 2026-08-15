package usecase

import (
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
)

type CheckoutItemInput struct {
	ProductID string
	Quantity  int
}

type CheckoutInput struct {
	UserID string
	Items  []CheckoutItemInput
}

type CheckoutOutput struct {
	Order *orderdomain.Order
}
