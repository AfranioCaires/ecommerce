package configuration

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var ErrInvalidConfiguration = errors.New("the payment service configuration is invalid")

type Configuration struct {
	ApplicationPort         string
	PostgreSQLDataSource    string
	RabbitMQURL             string
	RabbitMQCommandExchange string
	RabbitMQEventExchange   string
	RabbitMQPaymentQueue    string
	RabbitMQResultQueue     string
	OutboxInterval          time.Duration
	OutboxBatchSize         int
	MessageRetryLimit       int
	ShutdownTimeout         time.Duration
}

func Load() (*Configuration, error) {
	_ = godotenv.Load()

	outboxInterval, errorValue := positiveDuration("OUTBOX_INTERVAL", "250ms")
	if errorValue != nil {
		return nil, errorValue
	}
	shutdownTimeout, errorValue := positiveDuration("SHUTDOWN_TIMEOUT", "5s")
	if errorValue != nil {
		return nil, errorValue
	}
	outboxBatchSize, errorValue := positiveInteger("OUTBOX_BATCH_SIZE", "20")
	if errorValue != nil {
		return nil, errorValue
	}
	retryLimit, errorValue := positiveInteger("MESSAGE_RETRY_LIMIT", "5")
	if errorValue != nil {
		return nil, errorValue
	}

	configuration := &Configuration{
		ApplicationPort:         environmentValue("PAYMENT_APPLICATION_PORT", "3001"),
		PostgreSQLDataSource:    os.Getenv("PAYMENT_POSTGRESQL_DATA_SOURCE"),
		RabbitMQURL:             os.Getenv("RABBITMQ_URL"),
		RabbitMQCommandExchange: environmentValue("RABBITMQ_COMMAND_EXCHANGE", "ecommerce.commands"),
		RabbitMQEventExchange:   environmentValue("RABBITMQ_EVENT_EXCHANGE", "ecommerce.events"),
		RabbitMQPaymentQueue:    environmentValue("RABBITMQ_PAYMENT_QUEUE", "payment.requests"),
		RabbitMQResultQueue:     environmentValue("RABBITMQ_RESULT_QUEUE", "ecommerce.payment-results"),
		OutboxInterval:          outboxInterval,
		OutboxBatchSize:         outboxBatchSize,
		MessageRetryLimit:       retryLimit,
		ShutdownTimeout:         shutdownTimeout,
	}
	if configuration.ApplicationPort == "" || configuration.PostgreSQLDataSource == "" || configuration.RabbitMQURL == "" || configuration.RabbitMQCommandExchange == "" || configuration.RabbitMQEventExchange == "" || configuration.RabbitMQPaymentQueue == "" || configuration.RabbitMQResultQueue == "" {
		return nil, ErrInvalidConfiguration
	}
	return configuration, nil
}

func positiveDuration(name, fallback string) (time.Duration, error) {
	value, errorValue := time.ParseDuration(environmentValue(name, fallback))
	if errorValue != nil || value <= 0 {
		return 0, ErrInvalidConfiguration
	}
	return value, nil
}

func positiveInteger(name, fallback string) (int, error) {
	value, errorValue := strconv.Atoi(environmentValue(name, fallback))
	if errorValue != nil || value <= 0 {
		return 0, ErrInvalidConfiguration
	}
	return value, nil
}

func environmentValue(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
