package events

import "errors"

const (
	PaymentRequestedV1 = "payment.requested.v1"
	PaymentApprovedV1  = "payment.approved.v1"
	PaymentDeclinedV1  = "payment.declined.v1"
)

var ErrInvalidPaymentMessage = errors.New("the payment message is invalid.")

type PaymentRequested struct {
	OrderID     string `json:"order_id"`
	AmountCents int64  `json:"amount_cents"`
}

func (message PaymentRequested) Validate() error {
	if message.OrderID == "" || message.AmountCents <= 0 {
		return ErrInvalidPaymentMessage
	}
	return nil
}

type PaymentResult struct {
	OrderID     string `json:"order_id"`
	PaymentID   string `json:"payment_id"`
	AmountCents int64  `json:"amount_cents"`
	ReasonCode  string `json:"reason_code,omitempty"`
}

func (message PaymentResult) Validate() error {
	if message.OrderID == "" || message.PaymentID == "" || message.AmountCents <= 0 {
		return ErrInvalidPaymentMessage
	}
	return nil
}
