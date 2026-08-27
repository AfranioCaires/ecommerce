package configuration

import (
	"errors"
	"testing"
	"time"
)

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("PAYMENT_POSTGRESQL_DATA_SOURCE", "host=payment-database")
	t.Setenv("RABBITMQ_URL", "amqp://rabbitmq/")
}

func TestLoadUsesPaymentDefaults(t *testing.T) {
	setRequiredEnvironment(t)
	configuration, errorValue := Load()
	if errorValue != nil {
		t.Fatalf("Load() error = %v", errorValue)
	}
	if configuration.ApplicationPort != "3001" || configuration.OutboxInterval != 250*time.Millisecond || configuration.OutboxBatchSize != 20 || configuration.MessageRetryLimit != 5 || configuration.ShutdownTimeout != 5*time.Second {
		t.Fatalf("Load() = %#v", configuration)
	}
}

func TestLoadAcceptsCustomValues(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("PAYMENT_APPLICATION_PORT", "9001")
	t.Setenv("OUTBOX_INTERVAL", "1s")
	t.Setenv("OUTBOX_BATCH_SIZE", "40")
	t.Setenv("MESSAGE_RETRY_LIMIT", "3")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	configuration, errorValue := Load()
	if errorValue != nil {
		t.Fatalf("Load() error = %v", errorValue)
	}
	if configuration.ApplicationPort != "9001" || configuration.OutboxInterval != time.Second || configuration.OutboxBatchSize != 40 || configuration.MessageRetryLimit != 3 || configuration.ShutdownTimeout != 10*time.Second {
		t.Fatalf("Load() = %#v", configuration)
	}
}

func TestLoadRejectsMissingOrInvalidValues(t *testing.T) {
	t.Run("missing payment database", func(t *testing.T) {
		t.Setenv("PAYMENT_POSTGRESQL_DATA_SOURCE", "")
		t.Setenv("RABBITMQ_URL", "amqp://rabbitmq/")
		configuration, errorValue := Load()
		if configuration != nil || !errors.Is(errorValue, ErrInvalidConfiguration) {
			t.Fatalf("Load() = %#v, %v", configuration, errorValue)
		}
	})
	t.Run("missing broker", func(t *testing.T) {
		t.Setenv("PAYMENT_POSTGRESQL_DATA_SOURCE", "host=payment-database")
		t.Setenv("RABBITMQ_URL", "")
		configuration, errorValue := Load()
		if configuration != nil || !errors.Is(errorValue, ErrInvalidConfiguration) {
			t.Fatalf("Load() = %#v, %v", configuration, errorValue)
		}
	})
	t.Run("invalid batch", func(t *testing.T) {
		setRequiredEnvironment(t)
		t.Setenv("OUTBOX_BATCH_SIZE", "0")
		configuration, errorValue := Load()
		if configuration != nil || !errors.Is(errorValue, ErrInvalidConfiguration) {
			t.Fatalf("Load() = %#v, %v", configuration, errorValue)
		}
	})
}
