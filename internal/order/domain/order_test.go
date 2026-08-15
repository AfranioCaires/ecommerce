package domain

import (
	"errors"
	"testing"
	"time"
)

func TestOrder(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.FixedZone("test", -3*60*60))
	validItems := []OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1250, Quantity: 2}}

	t.Run("it should calculate snapshots and enforce payment transitions", func(t *testing.T) {
		order, errorValue := NewOrder("order-1", "user-1", []OrderItem{{ProductID: "product-1", ProductName: "Keyboard", UnitPriceCents: 1250, Quantity: 2}}, time.Now())
		if errorValue != nil || order.TotalAmountCents != 2500 || order.Status != OrderStatusPending {
			t.Fatalf("unexpected order: %#v, %v", order, errorValue)
		}
		if errorValue := order.StartPayment(time.Now()); errorValue != nil {
			t.Fatal(errorValue)
		}
		if errorValue := order.MarkAsPaid(time.Now()); errorValue != nil {
			t.Fatal(errorValue)
		}
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

	for name, testCase := range map[string]struct {
		orderID string
		userID  string
		items   []OrderItem
		wantErr error
	}{
		"empty order ID": {orderID: " ", userID: "user-1", items: validItems, wantErr: ErrEmptyOrderID},
		"empty user ID":  {orderID: "order-1", userID: " ", items: validItems, wantErr: ErrEmptyOrderUserID},
		"empty items":    {orderID: "order-1", userID: "user-1", wantErr: ErrEmptyOrderItems},
		"empty product name": {
			orderID: "order-1", userID: "user-1",
			items: []OrderItem{{ProductID: "product-1", ProductName: " ", UnitPriceCents: 100, Quantity: 1}}, wantErr: ErrInvalidOrderItem,
		},
		"invalid price": {
			orderID: "order-1", userID: "user-1",
			items: []OrderItem{{ProductID: "product-1", ProductName: "Product", UnitPriceCents: 0, Quantity: 1}}, wantErr: ErrInvalidOrderItem,
		},
		"invalid quantity": {
			orderID: "order-1", userID: "user-1",
			items: []OrderItem{{ProductID: "product-1", ProductName: "Product", UnitPriceCents: 100, Quantity: 0}}, wantErr: ErrInvalidOrderItem,
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOrder(testCase.orderID, testCase.userID, testCase.items, now)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("NewOrder() error = %v, want %v", err, testCase.wantErr)
			}
		})
	}

	t.Run("it should normalize timestamps and restore every valid status", func(t *testing.T) {
		for _, status := range []OrderStatus{OrderStatusPending, OrderStatusPaymentPending, OrderStatusPaid, OrderStatusCanceled} {
			updatedAt := now.Add(time.Hour)
			order, err := RestoreOrder("order-1", "user-1", validItems, 9999, status, now, updatedAt)
			if err != nil || order.Status != status || order.TotalAmountCents != 9999 || !order.CreatedAt.Equal(now.UTC()) || !order.UpdatedAt.Equal(updatedAt.UTC()) {
				t.Fatalf("RestoreOrder() = %#v, %v", order, err)
			}
		}
	})

	t.Run("it should propagate construction and status restoration errors", func(t *testing.T) {
		if _, err := RestoreOrder("", "user-1", validItems, 100, OrderStatusPending, now, now); !errors.Is(err, ErrEmptyOrderID) {
			t.Fatalf("RestoreOrder() error = %v", err)
		}
		if _, err := RestoreOrder("order-1", "user-1", validItems, 100, OrderStatus("UNKNOWN"), now, now); !errors.Is(err, ErrInvalidOrderStatus) {
			t.Fatalf("RestoreOrder() error = %v", err)
		}
	})

	t.Run("it should cancel pending orders and reject final transitions", func(t *testing.T) {
		order, _ := NewOrder("order-1", "user-1", validItems, now)
		updatedAt := now.Add(2 * time.Hour)
		if errorValue := order.Cancel(updatedAt); errorValue != nil {
			t.Fatal(errorValue)
		}
		if order.Status != OrderStatusCanceled || !order.UpdatedAt.Equal(updatedAt.UTC()) {
			t.Fatalf("unexpected canceled order: %#v", order)
		}
		if errorValue := order.StartPayment(updatedAt.Add(time.Hour)); !errors.Is(errorValue, ErrInvalidOrderTransition) {
			t.Fatalf("expected invalid transition, received %v", errorValue)
		}
	})

	t.Run("it should validate all statuses", func(t *testing.T) {
		for _, status := range []OrderStatus{OrderStatusPending, OrderStatusPaymentPending, OrderStatusPaid, OrderStatusCanceled} {
			if !status.IsValid() {
				t.Fatalf("expected %q to be valid", status)
			}
		}
		if OrderStatus("UNKNOWN").IsValid() {
			t.Fatal("expected unknown status to be invalid")
		}
	})
}
