package domain

import "errors"

var ErrInvalidProductStatus = errors.New("the product status is invalid.")

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusInactive ProductStatus = "INACTIVE"
)

func (productStatus ProductStatus) IsValid() bool {
	switch productStatus {
	case ProductStatusActive, ProductStatusInactive:
		return true
	default:
		return false
	}
}
