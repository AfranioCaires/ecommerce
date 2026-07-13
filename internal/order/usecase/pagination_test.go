package usecase

import (
	"errors"
	"testing"
)

func TestNewOrderPageRequest(t *testing.T) {
	t.Run("it should create valid order pagination", func(t *testing.T) {
		pageRequest, errorValue := NewOrderPageRequest(20, 10)
		if errorValue != nil || pageRequest.Limit != 20 || pageRequest.Offset != 10 {
			t.Fatalf("unexpected pagination: %#v, %v", pageRequest, errorValue)
		}
	})

	for description, input := range map[string]struct {
		limit  int
		offset int
	}{
		"it should reject a zero limit":      {limit: 0, offset: 0},
		"it should reject a limit above max": {limit: 101, offset: 0},
		"it should reject a negative offset": {limit: 20, offset: -1},
	} {
		t.Run(description, func(t *testing.T) {
			_, errorValue := NewOrderPageRequest(input.limit, input.offset)
			if !errors.Is(errorValue, ErrInvalidOrderPagination) {
				t.Fatalf("expected invalid pagination, received %v", errorValue)
			}
		})
	}
}
