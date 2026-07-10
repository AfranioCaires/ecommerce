package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidPaymentStatus = errors.New("the payment status is invalid.")
	ErrInvalidPaymentAmount = errors.New("the payment amount must be greater than zero.")
	ErrEmptyPaymentID       = errors.New("the payment ID must not be empty.")
	ErrEmptyPaymentOrderID  = errors.New("the payment order ID must not be empty.")
)

type PaymentStatus string

const (
	PaymentStatusApproved PaymentStatus = "APPROVED"
	PaymentStatusDeclined PaymentStatus = "DECLINED"
)

func (paymentStatus PaymentStatus) IsValid() bool {
	switch paymentStatus {
	case PaymentStatusApproved, PaymentStatusDeclined:
		return true
	default:
		return false
	}
}

type Payment struct {
	ID          string
	OrderID     string
	AmountCents int64
	Status      PaymentStatus
	CreatedAt   time.Time
}

func NewPayment(
	paymentID string,
	orderID string,
	amountCents int64,
	paymentStatus PaymentStatus,
	createdAt time.Time,
) (*Payment, error) {
	if strings.TrimSpace(paymentID) == "" {
		return nil, ErrEmptyPaymentID
	}

	if strings.TrimSpace(orderID) == "" {
		return nil, ErrEmptyPaymentOrderID
	}

	if amountCents <= 0 {
		return nil, ErrInvalidPaymentAmount
	}

	if !paymentStatus.IsValid() {
		return nil, ErrInvalidPaymentStatus
	}

	return &Payment{
		ID:          paymentID,
		OrderID:     orderID,
		AmountCents: amountCents,
		Status:      paymentStatus,
		CreatedAt:   createdAt.UTC(),
	}, nil
}
