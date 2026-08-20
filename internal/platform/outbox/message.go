package outbox

import "time"

type Message struct {
	ID, MessageType, RoutingKey string
	Payload                     []byte
	Attempts                    int
	NextAttemptAt               time.Time
	PublishedAt                 *time.Time
	LastError                   string
	CreatedAt                   time.Time
}

func NewMessage(id, messageType, routingKey string, payload []byte, createdAt time.Time) *Message {
	normalized := createdAt.UTC()
	return &Message{ID: id, MessageType: messageType, RoutingKey: routingKey, Payload: append([]byte(nil), payload...), NextAttemptAt: normalized, CreatedAt: normalized}
}
