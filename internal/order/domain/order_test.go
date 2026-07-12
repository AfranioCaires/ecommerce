package domain

import (
	"errors"
	"testing"
	"time"
)

func TestOrder(t *testing.T) {
	t.Run("it should calculate snapshots and mark an order as paid", func(t *testing.T) {
		order, errorValue := NewOrder("order-1", "user-1", []OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1250, Quantity: 2}}, time.Now())
		if errorValue != nil || order.TotalAmountCents != 2500 || order.Status != OrderStatusPending {
			t.Fatalf("unexpected order: %#v, %v", order, errorValue)
		}
		order.MarkAsPaid(time.Now())
		if order.Status != OrderStatusPaid {
			t.Fatalf("expected paid status, received %s", order.Status)
		}
	})

	t.Run("it should reject an invalid item", func(t *testing.T) {
		_, errorValue := NewOrder("order-1", "user-1", []OrderItem{{ProductID: "", ProductName: "Product", UnitPriceCents: 100, Quantity: 1}}, time.Now())
		if !errors.Is(errorValue, ErrInvalidOrderItem) {
			t.Fatalf("expected invalid item, received %v", errorValue)
		}
	})
}
