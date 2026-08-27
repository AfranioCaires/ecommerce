package paymentmessaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/platform/events"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
)

type processorFake struct {
	envelope events.Envelope
	request  events.PaymentRequested
	error    error
}

func (processor *processorFake) ProcessRequested(_ context.Context, envelope events.Envelope, request events.PaymentRequested) error {
	processor.envelope, processor.request = envelope, request
	return processor.error
}

func TestPaymentRequestHandler(t *testing.T) {
	now := time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
	validEnvelope, _ := events.NewEnvelope("message-1", events.PaymentRequestedV1, "saga-1", "correlation-1", now, events.PaymentRequested{OrderID: "order-1", AmountCents: 1000})
	validBody, _ := json.Marshal(validEnvelope)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("valid request", func(t *testing.T) {
		processor := &processorFake{}
		handler := NewPaymentRequestHandler(processor, logger)
		if errorValue := handler.Handle(context.Background(), validBody); errorValue != nil || processor.request.OrderID != "order-1" || processor.envelope.MessageID != "message-1" {
			t.Fatalf("Handle() processor=%#v error=%v", processor, errorValue)
		}
	})

	t.Run("processing failure stays retryable", func(t *testing.T) {
		expectedError := errors.New("database unavailable")
		handler := NewPaymentRequestHandler(&processorFake{error: expectedError}, logger)
		errorValue := handler.Handle(context.Background(), validBody)
		var permanentError platformmessaging.PermanentError
		if !errors.Is(errorValue, expectedError) || errors.As(errorValue, &permanentError) {
			t.Fatalf("Handle() error = %v", errorValue)
		}
	})

	for _, testCase := range []struct {
		name string
		body []byte
	}{
		{"malformed envelope", []byte("{")},
		{"unknown message", messageBody(t, "other.v1", events.PaymentRequested{OrderID: "order-1", AmountCents: 1000}, now)},
		{"invalid payload", messageBody(t, events.PaymentRequestedV1, events.PaymentRequested{OrderID: "", AmountCents: 0}, now)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewPaymentRequestHandler(&processorFake{}, logger)
			errorValue := handler.Handle(context.Background(), testCase.body)
			var permanentError platformmessaging.PermanentError
			if !errors.As(errorValue, &permanentError) {
				t.Fatalf("Handle() error = %v, want permanent", errorValue)
			}
		})
	}
}

func messageBody(t *testing.T, messageType string, payload any, at time.Time) []byte {
	t.Helper()
	envelope, errorValue := events.NewEnvelope("message-2", messageType, "saga-2", "correlation-2", at, payload)
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	body, errorValue := json.Marshal(envelope)
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return body
}
