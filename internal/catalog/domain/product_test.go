package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewProduct(t *testing.T) {
	t.Run("it should create a normalized active product", func(t *testing.T) {
		createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.FixedZone("test", -10800))
		product, errorValue := NewProduct("product-1", " Keyboard ", " Mechanical ", 10000, createdAt)
		if errorValue != nil || product.Name != "Keyboard" || product.Description != "Mechanical" || product.Status != ProductStatusActive || product.CreatedAt.Location() != time.UTC {
			t.Fatalf("unexpected product: %#v, %v", product, errorValue)
		}
	})

	for description, input := range map[string]struct {
		productID string
		name      string
		price     int64
		expected  error
	}{
		"it should reject an empty product ID":   {"", "Product", 100, ErrEmptyProductID},
		"it should reject an empty product name": {"product-1", " ", 100, ErrEmptyProductName},
		"it should reject a nonpositive price":   {"product-1", "Product", 0, ErrInvalidProductPrice},
	} {
		t.Run(description, func(t *testing.T) {
			_, errorValue := NewProduct(input.productID, input.name, "", input.price, time.Now())
			if !errors.Is(errorValue, input.expected) {
				t.Fatalf("expected %v, received %v", input.expected, errorValue)
			}
		})
	}
}
