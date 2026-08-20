package outbox

import (
	"context"
	"time"
)

type Repository interface {
	Save(context.Context, *Message) error
	Pending(context.Context, time.Time, int) ([]*Message, error)
	MarkPublished(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, time.Time, string) error
}
