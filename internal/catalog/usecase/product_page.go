package usecase

import (
	"math"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
)

type ProductPage struct {
	Products   []*domain.Product
	PageNumber int
	PageSize   int
	TotalItems int64
	TotalPages int
}

func NewProductPage(
	products []*domain.Product,
	pageRequest ProductPageRequest,
	totalItems int64,
) ProductPage {
	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(
			float64(totalItems) / float64(pageRequest.PageSize),
		))
	}

	return ProductPage{
		Products:   products,
		PageNumber: pageRequest.PageNumber,
		PageSize:   pageRequest.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
