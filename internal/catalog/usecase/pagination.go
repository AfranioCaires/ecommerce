package usecase

import "errors"

var ErrInvalidPagination = errors.New("the pagination values are invalid.")

const (
	DefaultPageNumber = 1
	DefaultPageSize   = 20
	MaximumPageSize   = 100
)

type ProductPageRequest struct {
	PageNumber int
	PageSize   int
}

func NewProductPageRequest(
	pageNumber int,
	pageSize int,
) (ProductPageRequest, error) {
	if pageNumber < 1 || pageSize < 1 || pageSize > MaximumPageSize {
		return ProductPageRequest{}, ErrInvalidPagination
	}

	return ProductPageRequest{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}, nil
}
