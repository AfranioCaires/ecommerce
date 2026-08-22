package configuration

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var ErrInvalidConfiguration = errors.New("the application configuration is invalid.")

var jsonWebTokenSecretFallback = "RED-DEAD-REDEMPTION-2"

type Configuration struct {
	ApplicationPort         string
	ApplicationEnvironment  string
	PostgreSQLDataSource    string
	JSONWebTokenSecret      string
	JSONWebTokenIssuer      string
	JSONWebTokenLifetime    time.Duration
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

	jsonWebTokenLifetime, errorValue := time.ParseDuration(
		environmentValue("JSON_WEB_TOKEN_LIFETIME", "15m"),
	)
	if errorValue != nil {
		return nil, ErrInvalidConfiguration
	}
	outboxInterval, errorValue := time.ParseDuration(environmentValue("OUTBOX_INTERVAL", "250ms"))
	if errorValue != nil || outboxInterval <= 0 {
		return nil, ErrInvalidConfiguration
	}
	shutdownTimeout, errorValue := time.ParseDuration(environmentValue("SHUTDOWN_TIMEOUT", "5s"))
	if errorValue != nil || shutdownTimeout <= 0 {
		return nil, ErrInvalidConfiguration
	}
	outboxBatchSize, errorValue := strconv.Atoi(environmentValue("OUTBOX_BATCH_SIZE", "20"))
	if errorValue != nil || outboxBatchSize <= 0 {
		return nil, ErrInvalidConfiguration
	}
	retryLimit, errorValue := strconv.Atoi(environmentValue("MESSAGE_RETRY_LIMIT", "5"))
	if errorValue != nil || retryLimit <= 0 {
		return nil, ErrInvalidConfiguration
	}

	applicationConfiguration := &Configuration{
		ApplicationPort:         environmentValue("APPLICATION_PORT", "3000"),
		ApplicationEnvironment:  environmentValue("APPLICATION_ENVIRONMENT", "development"),
		PostgreSQLDataSource:    environmentValue("POSTGRESQL_DATA_SOURCE", "host=localhost port=5432 user=afraniocaires password=postgres dbname=ecommerce sslmode=disable"),
		JSONWebTokenSecret:      environmentValue("JSON_WEB_TOKEN_SECRET", jsonWebTokenSecretFallback),
		JSONWebTokenIssuer:      environmentValue("JSON_WEB_TOKEN_ISSUER", "afranio"),
		JSONWebTokenLifetime:    jsonWebTokenLifetime,
		RabbitMQURL:             environmentValue("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQCommandExchange: environmentValue("RABBITMQ_COMMAND_EXCHANGE", "ecommerce.commands"),
		RabbitMQEventExchange:   environmentValue("RABBITMQ_EVENT_EXCHANGE", "ecommerce.events"),
		RabbitMQPaymentQueue:    environmentValue("RABBITMQ_PAYMENT_QUEUE", "payment.requests"),
		RabbitMQResultQueue:     environmentValue("RABBITMQ_RESULT_QUEUE", "ecommerce.payment-results"),
		OutboxInterval:          outboxInterval,
		OutboxBatchSize:         outboxBatchSize,
		MessageRetryLimit:       retryLimit,
		ShutdownTimeout:         shutdownTimeout,
	}

	if applicationConfiguration.JSONWebTokenSecret == "" || applicationConfiguration.RabbitMQURL == "" || applicationConfiguration.RabbitMQCommandExchange == "" || applicationConfiguration.RabbitMQEventExchange == "" || applicationConfiguration.RabbitMQPaymentQueue == "" || applicationConfiguration.RabbitMQResultQueue == "" {
		return nil, ErrInvalidConfiguration
	}

	return applicationConfiguration, nil
}

func environmentValue(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	return value
}
