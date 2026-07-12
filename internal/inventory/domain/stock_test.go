package domain

import (
	"errors"
	"testing"
	"time"
)

func TestStock(t *testing.T) {
	t.Run("it should reserve and release stock", func(t *testing.T) {
		stock, _ := NewStock("product-1", 10, time.Now())
		if errorValue := stock.Reserve(4, time.Now()); errorValue != nil {
			t.Fatal(errorValue)
		}
		if errorValue := stock.Release(2, time.Now()); errorValue != nil {
			t.Fatal(errorValue)
		}
		if stock.Quantity != 8 {
			t.Fatalf("expected quantity 8, received %d", stock.Quantity)
		}
	})

	t.Run("it should reject insufficient stock", func(t *testing.T) {
		stock, _ := NewStock("product-1", 1, time.Now())
		errorValue := stock.Reserve(2, time.Now())
		if !errors.Is(errorValue, ErrInsufficientStock) {
			t.Fatalf("expected insufficient stock, received %v", errorValue)
		}
	})
}
