package outbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type publisherRepository struct {
	message           *Message
	published, failed bool
}

func (repository *publisherRepository) Save(context.Context, *Message) error { return nil }
func (repository *publisherRepository) Pending(context.Context, time.Time, int) ([]*Message, error) {
	return []*Message{repository.message}, nil
}
func (repository *publisherRepository) MarkPublished(context.Context, string, time.Time) error {
	repository.published = true
	return nil
}
func (repository *publisherRepository) MarkFailed(context.Context, string, time.Time, string) error {
	repository.failed = true
	return nil
}

type publisherBroker struct{ errorValue error }

func (broker publisherBroker) Publish(context.Context, string, []byte, map[string]any) error {
	return broker.errorValue
}

func TestPublisherMarksSuccessAndFailure(t *testing.T) {
	now := time.Now()
	for _, testCase := range []struct {
		name              string
		brokerError       error
		published, failed bool
	}{
		{"success", nil, true, false}, {"failure", errors.New("broker failed"), false, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &publisherRepository{message: NewMessage("message-1", "type", "route", []byte("{}"), now)}
			publisher := NewPublisher(repository, publisherBroker{testCase.brokerError}, time.Second, 10, func() time.Time { return now }, nil)
			if errorValue := publisher.PublishPending(context.Background()); errorValue != nil {
				t.Fatal(errorValue)
			}
			if repository.published != testCase.published || repository.failed != testCase.failed {
				t.Fatalf("published=%v failed=%v", repository.published, repository.failed)
			}
		})
	}
}
