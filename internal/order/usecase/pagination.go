package usecase

import "errors"

var ErrInvalidOrderPagination = errors.New("the order pagination values are invalid.")

const (
	DefaultOrderLimit  = 20
	DefaultOrderOffset = 0
	MaximumOrderLimit  = 100
)

type OrderPageRequest struct {
	Limit  int
	Offset int
}

func NewOrderPageRequest(limit int, offset int) (OrderPageRequest, error) {
	if limit < 1 || limit > MaximumOrderLimit || offset < 0 {
		return OrderPageRequest{}, ErrInvalidOrderPagination
	}

	return OrderPageRequest{Limit: limit, Offset: offset}, nil
}
