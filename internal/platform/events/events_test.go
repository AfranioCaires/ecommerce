package events

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestEnvelopeAndPaymentContracts(t *testing.T) {
	envelope, errorValue := NewEnvelope("message-1", PaymentRequestedV1, "saga-1", "correlation-1", time.Now(), PaymentRequested{OrderID: "order-1", AmountCents: 100})
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	decoded, errorValue := Decode(mustJSON(envelope))
	payload, payloadError := PayloadAs[PaymentRequested](decoded)
	if errorValue != nil || payloadError != nil || payload.Validate() != nil || payload.OrderID != "order-1" {
		t.Fatalf("decoded=%#v payload=%#v errors=%v/%v", decoded, payload, errorValue, payloadError)
	}
	if _, errorValue := Decode([]byte("{")); !errors.Is(errorValue, ErrInvalidEnvelope) {
		t.Fatalf("expected invalid envelope, received %v", errorValue)
	}
	if errorValue := (PaymentResult{}).Validate(); !errors.Is(errorValue, ErrInvalidPaymentMessage) {
		t.Fatalf("expected invalid payment result, received %v", errorValue)
	}
}

func mustJSON(value any) []byte {
	data, errorValue := json.Marshal(value)
	if errorValue != nil {
		panic(errorValue)
	}
	return data
}
