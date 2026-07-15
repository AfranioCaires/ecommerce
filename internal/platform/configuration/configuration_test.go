package configuration

import (
	"errors"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Run("custom values", func(t *testing.T) {
		t.Setenv("APPLICATION_PORT", "9000")
		t.Setenv("APPLICATION_ENVIRONMENT", "test")
		t.Setenv("POSTGRESQL_DATA_SOURCE", "host=test")
		t.Setenv("JSON_WEB_TOKEN_SECRET", "secret")
		t.Setenv("JSON_WEB_TOKEN_ISSUER", "issuer")
		t.Setenv("JSON_WEB_TOKEN_LIFETIME", "1h")

		configuration, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if configuration.ApplicationPort != "9000" ||
			configuration.ApplicationEnvironment != "test" ||
			configuration.PostgreSQLDataSource != "host=test" ||
			configuration.JSONWebTokenSecret != "secret" ||
			configuration.JSONWebTokenIssuer != "issuer" ||
			configuration.JSONWebTokenLifetime != time.Hour {
			t.Fatalf("Load() = %#v", configuration)
		}
	})

	t.Run("invalid lifetime", func(t *testing.T) {
		t.Setenv("JSON_WEB_TOKEN_LIFETIME", "invalid")
		configuration, err := Load()
		if configuration != nil || !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("Load() = %#v, %v", configuration, err)
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		originalFallback := jsonWebTokenSecretFallback
		jsonWebTokenSecretFallback = ""
		t.Cleanup(func() { jsonWebTokenSecretFallback = originalFallback })
		t.Setenv("JSON_WEB_TOKEN_SECRET", "")

		configuration, err := Load()
		if configuration != nil || !errors.Is(err, ErrInvalidConfiguration) {
			t.Fatalf("Load() = %#v, %v", configuration, err)
		}
	})
}

func TestEnvironmentValue(t *testing.T) {
	const name = "ECOMMERCE_CONFIGURATION_TEST"
	t.Setenv(name, "")
	if value := environmentValue(name, "fallback"); value != "fallback" {
		t.Fatalf("environmentValue() = %q", value)
	}
	t.Setenv(name, "value")
	if value := environmentValue(name, "fallback"); value != "value" {
		t.Fatalf("environmentValue() = %q", value)
	}
}
