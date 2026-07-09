package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyProductID      = errors.New("the product ID must not be empty.")
	ErrEmptyProductName    = errors.New("the product name must not be empty.")
	ErrInvalidProductPrice = errors.New("the product price must be greater than zero.")
	ErrProductNotFound     = errors.New("the product was not found.")
)

type Product struct {
	ID          string
	Name        string
	Description string
	PriceCents  int64
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProduct(
	productID string,
	name string,
	description string,
	priceCents int64,
	createdAt time.Time,
) (*Product, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, ErrEmptyProductID
	}

	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyProductName
	}

	if priceCents <= 0 {
		return nil, ErrInvalidProductPrice
	}

	normalizedTime := createdAt.UTC()

	return &Product{
		ID:          productID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		PriceCents:  priceCents,
		Status:      ProductStatusActive,
		CreatedAt:   normalizedTime,
		UpdatedAt:   normalizedTime,
	}, nil
}

func RestoreProduct(
	productID string,
	name string,
	description string,
	priceCents int64,
	productStatus ProductStatus,
	createdAt time.Time,
	updatedAt time.Time,
) (*Product, error) {
	product, errorValue := NewProduct(
		productID,
		name,
		description,
		priceCents,
		createdAt,
	)
	if errorValue != nil {
		return nil, errorValue
	}

	if !productStatus.IsValid() {
		return nil, ErrInvalidProductStatus
	}

	product.Status = productStatus
	product.UpdatedAt = updatedAt.UTC()

	return product, nil
}
