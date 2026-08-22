package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/afraniocaires/ecommerce/internal/platform/events"
)

type Config struct {
	URL, CommandExchange, EventExchange, PaymentQueue, ResultQueue string
	RetryLimit                                                     int
}

type RabbitMQ struct {
	connection *amqp.Connection
	config     Config
}

func Dial(config Config) (*RabbitMQ, error) {
	connection, errorValue := amqp.Dial(config.URL)
	if errorValue != nil {
		return nil, fmt.Errorf("connect RabbitMQ: %w", errorValue)
	}
	broker := &RabbitMQ{connection: connection, config: config}
	if errorValue := broker.declareTopology(); errorValue != nil {
		_ = connection.Close()
		return nil, errorValue
	}
	return broker, nil
}

func (broker *RabbitMQ) Close() error { return broker.connection.Close() }

func (broker *RabbitMQ) declareTopology() error {
	channel, errorValue := broker.connection.Channel()
	if errorValue != nil {
		return errorValue
	}
	defer channel.Close()
	for _, exchange := range []string{broker.config.CommandExchange, broker.config.EventExchange, "ecommerce.dead"} {
		if errorValue := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); errorValue != nil {
			return errorValue
		}
	}
	definitions := []struct{ queue, exchange, key string }{
		{broker.config.PaymentQueue, broker.config.CommandExchange, "payment.requested"},
		{broker.config.ResultQueue, broker.config.EventExchange, "payment.*"},
	}
	for _, definition := range definitions {
		if _, errorValue := channel.QueueDeclare(definition.queue, true, false, false, false, nil); errorValue != nil {
			return errorValue
		}
		if errorValue := channel.QueueBind(definition.queue, definition.key, definition.exchange, false, nil); errorValue != nil {
			return errorValue
		}
		retryQueue := definition.queue + ".retry"
		if _, errorValue := channel.QueueDeclare(retryQueue, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange":    definition.exchange,
			"x-dead-letter-routing-key": definition.key,
		}); errorValue != nil {
			return errorValue
		}
		deadQueue := definition.queue + ".dead"
		if _, errorValue := channel.QueueDeclare(deadQueue, true, false, false, false, nil); errorValue != nil {
			return errorValue
		}
		if errorValue := channel.QueueBind(deadQueue, definition.queue, "ecommerce.dead", false, nil); errorValue != nil {
			return errorValue
		}
	}
	return nil
}

type Publisher interface {
	Publish(context.Context, string, []byte, map[string]any) error
}

type AMQPPublisher struct {
	channel *amqp.Channel
	config  Config
}

func (broker *RabbitMQ) NewPublisher() (*AMQPPublisher, error) {
	channel, errorValue := broker.connection.Channel()
	if errorValue != nil {
		return nil, errorValue
	}
	if errorValue := channel.Confirm(false); errorValue != nil {
		_ = channel.Close()
		return nil, errorValue
	}
	return &AMQPPublisher{channel: channel, config: broker.config}, nil
}

func (publisher *AMQPPublisher) Close() error { return publisher.channel.Close() }

func (publisher *AMQPPublisher) Publish(applicationContext context.Context, routingKey string, body []byte, headers map[string]any) error {
	envelope, errorValue := events.Decode(body)
	if errorValue != nil {
		return errorValue
	}
	exchange := publisher.config.EventExchange
	if routingKey == "payment.requested" {
		exchange = publisher.config.CommandExchange
	}
	deferredConfirmation, errorValue := publisher.channel.PublishWithDeferredConfirmWithContext(applicationContext, exchange, routingKey, true, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, ContentType: "application/json", MessageId: envelope.MessageID,
		Timestamp: envelope.OccurredAt, Headers: amqp.Table(headers), Body: body,
	})
	if errorValue != nil {
		return errorValue
	}
	if deferredConfirmation == nil {
		return errors.New("RabbitMQ publisher confirmation is unavailable")
	}
	confirmed, errorValue := deferredConfirmation.WaitContext(applicationContext)
	if errorValue != nil {
		return errorValue
	}
	if !confirmed {
		return errors.New("RabbitMQ rejected the publication")
	}
	return nil
}

type PermanentError struct{ Err error }

func (errorValue PermanentError) Error() string { return errorValue.Err.Error() }
func (errorValue PermanentError) Unwrap() error { return errorValue.Err }
func Permanent(errorValue error) error          { return PermanentError{Err: errorValue} }

type Handler func(context.Context, []byte) error

func retryDelay(attempt int) time.Duration {
	delay := time.Duration(1<<min(attempt, 6)) * time.Second
	if delay > time.Minute {
		return time.Minute
	}
	return delay
}
