package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyStockProductID  = errors.New("the stock product ID must not be empty.")
	ErrInvalidStockQuantity = errors.New("the stock quantity must not be negative.")
	ErrInsufficientStock    = errors.New("the available stock is insufficient.")
	ErrStockNotFound        = errors.New("the stock record was not found.")
)

type Stock struct {
	ProductID string
	Quantity  int
	UpdatedAt time.Time
}

func NewStock(
	productID string,
	quantity int,
	updatedAt time.Time,
) (*Stock, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, ErrEmptyStockProductID
	}

	if quantity < 0 {
		return nil, ErrInvalidStockQuantity
	}

	return &Stock{
		ProductID: productID,
		Quantity:  quantity,
		UpdatedAt: updatedAt.UTC(),
	}, nil
}

func (stock *Stock) Reserve(
	quantity int,
	updatedAt time.Time,
) error {
	if quantity <= 0 {
		return ErrInvalidStockQuantity
	}

	if stock.Quantity < quantity {
		return ErrInsufficientStock
	}

	stock.Quantity -= quantity
	stock.UpdatedAt = updatedAt.UTC()

	return nil
}

func (stock *Stock) Release(
	quantity int,
	updatedAt time.Time,
) error {
	if quantity <= 0 {
		return ErrInvalidStockQuantity
	}

	stock.Quantity += quantity
	stock.UpdatedAt = updatedAt.UTC()

	return nil
}
