package usecase

import (
	"context"
	"testing"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
)

func TestSimulatedPaymentGateway(t *testing.T) {
	t.Run("it should decline an amount ending in thirteen cents", func(t *testing.T) {
		paymentGateway := NewSimulatedPaymentGateway()
		paymentStatus, errorValue := paymentGateway.Authorize(context.Background(), "order-1", 1013)
		if errorValue != nil || paymentStatus != domain.PaymentStatusDeclined {
			t.Fatalf("expected declined payment, received %s and %v", paymentStatus, errorValue)
		}
	})

	t.Run("it should approve another amount", func(t *testing.T) {
		paymentGateway := NewSimulatedPaymentGateway()
		paymentStatus, errorValue := paymentGateway.Authorize(context.Background(), "order-1", 1014)
		if errorValue != nil || paymentStatus != domain.PaymentStatusApproved {
			t.Fatalf("expected approved payment, received %s and %v", paymentStatus, errorValue)
		}
	})
}
