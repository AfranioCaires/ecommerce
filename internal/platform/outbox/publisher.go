package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/afraniocaires/ecommerce/internal/platform/messaging"
)

type Publisher struct {
	repository  Repository
	broker      messaging.Publisher
	interval    time.Duration
	batchSize   int
	currentTime func() time.Time
	logger      *slog.Logger
}

func NewPublisher(repository Repository, broker messaging.Publisher, interval time.Duration, batchSize int, currentTime func() time.Time, logger *slog.Logger) *Publisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Publisher{repository: repository, broker: broker, interval: interval, batchSize: batchSize, currentTime: currentTime, logger: logger}
}

func (publisher *Publisher) Run(applicationContext context.Context) error {
	ticker := time.NewTicker(publisher.interval)
	defer ticker.Stop()
	for {
		if errorValue := publisher.PublishPending(applicationContext); errorValue != nil && applicationContext.Err() == nil {
			publisher.logger.Error("Outbox polling failed.", "operation", "outbox.publish", "result", "failed", "error", errorValue)
		}
		select {
		case <-applicationContext.Done():
			return applicationContext.Err()
		case <-ticker.C:
		}
	}
}

func (publisher *Publisher) PublishPending(applicationContext context.Context) error {
	now := publisher.currentTime().UTC()
	messages, errorValue := publisher.repository.Pending(applicationContext, now, publisher.batchSize)
	if errorValue != nil {
		return errorValue
	}
	for _, message := range messages {
		errorValue := publisher.broker.Publish(applicationContext, message.RoutingKey, message.Payload, map[string]any{"x-outbox-attempt": message.Attempts + 1})
		if errorValue == nil {
			if errorValue = publisher.repository.MarkPublished(applicationContext, message.ID, publisher.currentTime()); errorValue != nil {
				return errorValue
			}
			publisher.logger.Info("Outbox message published.", "operation", "outbox.publish", "result", "success", "message_id", message.ID, "message_type", message.MessageType, "routing_key", message.RoutingKey)
			continue
		}
		delay := time.Duration(1<<min(message.Attempts, 6)) * time.Second
		if delay > time.Minute {
			delay = time.Minute
		}
		if repositoryError := publisher.repository.MarkFailed(applicationContext, message.ID, publisher.currentTime().Add(delay), errorValue.Error()); repositoryError != nil {
			return repositoryError
		}
		publisher.logger.Warn("Outbox publication scheduled for retry.", "operation", "outbox.publish", "result", "retry", "message_id", message.ID, "message_type", message.MessageType, "attempt", message.Attempts+1, "error", errorValue)
	}
	return nil
}
