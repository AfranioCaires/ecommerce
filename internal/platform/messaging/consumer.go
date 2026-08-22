package messaging

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	channel *amqp.Channel
	queue   string
	config  Config
}

func (broker *RabbitMQ) NewConsumer(queue string) (*Consumer, error) {
	channel, errorValue := broker.connection.Channel()
	if errorValue != nil {
		return nil, errorValue
	}
	if errorValue := channel.Qos(1, 0, false); errorValue != nil {
		_ = channel.Close()
		return nil, errorValue
	}
	if errorValue := channel.Confirm(false); errorValue != nil {
		_ = channel.Close()
		return nil, errorValue
	}
	return &Consumer{channel: channel, queue: queue, config: broker.config}, nil
}

func (consumer *Consumer) Close() error { return consumer.channel.Close() }

func (consumer *Consumer) Run(applicationContext context.Context, handler Handler) error {
	deliveries, errorValue := consumer.channel.Consume(consumer.queue, "", false, false, false, false, nil)
	if errorValue != nil {
		return errorValue
	}
	for {
		select {
		case <-applicationContext.Done():
			return applicationContext.Err()
		case delivery, available := <-deliveries:
			if !available {
				return errors.New("RabbitMQ delivery channel closed")
			}
			if errorValue := handler(applicationContext, delivery.Body); errorValue == nil {
				if errorValue := delivery.Ack(false); errorValue != nil {
					return errorValue
				}
				continue
			} else if errorValue := consumer.handleFailure(applicationContext, delivery, errorValue); errorValue != nil {
				return errorValue
			}
		}
	}
}

func (consumer *Consumer) handleFailure(applicationContext context.Context, delivery amqp.Delivery, handlerError error) error {
	attempt := headerInteger(delivery.Headers, "x-retry-count") + 1
	var permanentError PermanentError
	dead := errors.As(handlerError, &permanentError) || attempt > consumer.config.RetryLimit
	exchange, routingKey := "ecommerce.dead", consumer.queue
	expiration := ""
	if !dead {
		exchange, routingKey = "", consumer.queue+".retry"
		expiration = strconv.FormatInt(retryDelay(attempt).Milliseconds(), 10)
	}
	headers := delivery.Headers
	if headers == nil {
		headers = amqp.Table{}
	}
	headers["x-retry-count"] = int32(attempt)
	if !dead {
		headers["x-retry-delay"] = retryDelay(attempt).String()
	}
	confirmation, errorValue := consumer.channel.PublishWithDeferredConfirmWithContext(applicationContext, exchange, routingKey, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: delivery.ContentType, Headers: headers, Expiration: expiration, Body: delivery.Body})
	if errorValue != nil {
		return errorValue
	}
	if confirmation == nil {
		return fmt.Errorf("retry confirmation unavailable: %w", handlerError)
	}
	confirmed, errorValue := confirmation.WaitContext(applicationContext)
	if errorValue != nil || !confirmed {
		return fmt.Errorf("confirm retry publication: %w", errorValue)
	}
	return delivery.Ack(false)
}

func headerInteger(headers amqp.Table, name string) int {
	value, available := headers[name]
	if !available {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}
