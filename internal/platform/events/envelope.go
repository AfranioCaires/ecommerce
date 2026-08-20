package events

import (
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidEnvelope = errors.New("the message envelope is invalid.")

type Envelope struct {
	MessageID     string          `json:"message_id"`
	MessageType   string          `json:"message_type"`
	SagaID        string          `json:"saga_id"`
	CorrelationID string          `json:"correlation_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

func NewEnvelope(messageID, messageType, sagaID, correlationID string, occurredAt time.Time, payload any) (Envelope, error) {
	if messageID == "" || messageType == "" || sagaID == "" || correlationID == "" || occurredAt.IsZero() {
		return Envelope{}, ErrInvalidEnvelope
	}
	payloadJSON, errorValue := json.Marshal(payload)
	if errorValue != nil {
		return Envelope{}, errorValue
	}
	return Envelope{MessageID: messageID, MessageType: messageType, SagaID: sagaID, CorrelationID: correlationID, OccurredAt: occurredAt.UTC(), Payload: payloadJSON}, nil
}

func Decode(body []byte) (Envelope, error) {
	var envelope Envelope
	if errorValue := json.Unmarshal(body, &envelope); errorValue != nil {
		return Envelope{}, ErrInvalidEnvelope
	}
	if envelope.MessageID == "" || envelope.MessageType == "" || envelope.SagaID == "" || envelope.CorrelationID == "" || envelope.OccurredAt.IsZero() || len(envelope.Payload) == 0 {
		return Envelope{}, ErrInvalidEnvelope
	}
	return envelope, nil
}

func PayloadAs[T any](envelope Envelope) (T, error) {
	var payload T
	if errorValue := json.Unmarshal(envelope.Payload, &payload); errorValue != nil {
		return payload, ErrInvalidEnvelope
	}
	return payload, nil
}
