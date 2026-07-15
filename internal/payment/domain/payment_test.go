package domain

import (
	"errors"
	"testing"
	"time"
)

func TestPaymentStatusIsValid(t *testing.T) {
	for _, testCase := range []struct {
		status PaymentStatus
		want   bool
	}{
		{PaymentStatusApproved, true},
		{PaymentStatusDeclined, true},
		{PaymentStatus("UNKNOWN"), false},
	} {
		if got := testCase.status.IsValid(); got != testCase.want {
			t.Fatalf("PaymentStatus(%q).IsValid() = %t", testCase.status, got)
		}
	}
}

func TestNewPayment(t *testing.T) {
	createdAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.FixedZone("test", -3*60*60))
	testCases := []struct {
		name      string
		id        string
		orderID   string
		amount    int64
		status    PaymentStatus
		wantError error
	}{
		{"approved", "payment-1", "order-1", 100, PaymentStatusApproved, nil},
		{"declined", "payment-1", "order-1", 100, PaymentStatusDeclined, nil},
		{"empty payment id", " ", "order-1", 100, PaymentStatusApproved, ErrEmptyPaymentID},
		{"empty order id", "payment-1", " ", 100, PaymentStatusApproved, ErrEmptyPaymentOrderID},
		{"invalid amount", "payment-1", "order-1", 0, PaymentStatusApproved, ErrInvalidPaymentAmount},
		{"invalid status", "payment-1", "order-1", 100, PaymentStatus("UNKNOWN"), ErrInvalidPaymentStatus},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			payment, errorValue := NewPayment(testCase.id, testCase.orderID, testCase.amount, testCase.status, createdAt)
			if !errors.Is(errorValue, testCase.wantError) {
				t.Fatalf("NewPayment() error = %v, want %v", errorValue, testCase.wantError)
			}
			if testCase.wantError != nil {
				if payment != nil {
					t.Fatalf("NewPayment() payment = %#v", payment)
				}
				return
			}
			if payment.ID != testCase.id || payment.OrderID != testCase.orderID || payment.AmountCents != testCase.amount || payment.Status != testCase.status || !payment.CreatedAt.Equal(createdAt) || payment.CreatedAt.Location() != time.UTC {
				t.Fatalf("NewPayment() = %#v", payment)
			}
		})
	}
}
