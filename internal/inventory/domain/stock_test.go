package domain

import (
	"errors"
	"testing"
	"time"
)

func TestStock(t *testing.T) {
	t.Run("it should reject an empty product ID", func(t *testing.T) {
		stock, errorValue := NewStock(" ", 1, time.Now())
		if stock != nil || !errors.Is(errorValue, ErrEmptyStockProductID) {
			t.Fatalf("NewStock() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("it should reject a negative initial quantity", func(t *testing.T) {
		stock, errorValue := NewStock("product-1", -1, time.Now())
		if stock != nil || !errors.Is(errorValue, ErrInvalidStockQuantity) {
			t.Fatalf("NewStock() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("it should create stock with an UTC update time", func(t *testing.T) {
		updatedAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.FixedZone("test", -3*60*60))
		stock, errorValue := NewStock("product-1", 3, updatedAt)
		if errorValue != nil || stock.ProductID != "product-1" || stock.Quantity != 3 || !stock.UpdatedAt.Equal(updatedAt) || stock.UpdatedAt.Location() != time.UTC {
			t.Fatalf("NewStock() = %#v, %v", stock, errorValue)
		}
	})

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

	t.Run("it should reject invalid reservation quantities", func(t *testing.T) {
		stock, _ := NewStock("product-1", 1, time.Now())
		for _, quantity := range []int{0, -1} {
			if errorValue := stock.Reserve(quantity, time.Now()); !errors.Is(errorValue, ErrInvalidStockQuantity) {
				t.Fatalf("Reserve(%d) error = %v", quantity, errorValue)
			}
		}
	})

	t.Run("it should reject invalid release quantities", func(t *testing.T) {
		stock, _ := NewStock("product-1", 1, time.Now())
		for _, quantity := range []int{0, -1} {
			if errorValue := stock.Release(quantity, time.Now()); !errors.Is(errorValue, ErrInvalidStockQuantity) {
				t.Fatalf("Release(%d) error = %v", quantity, errorValue)
			}
		}
	})
}
