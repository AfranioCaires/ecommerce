package configuration

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var ErrInvalidConfiguration = errors.New("the application configuration is invalid.")

var jsonWebTokenSecretFallback = "RED-DEAD-REDEMPTION-2"

type Configuration struct {
	ApplicationPort        string
	ApplicationEnvironment string
	PostgreSQLDataSource   string
	JSONWebTokenSecret     string
	JSONWebTokenIssuer     string
	JSONWebTokenLifetime   time.Duration
}

func Load() (*Configuration, error) {
	_ = godotenv.Load()

	jsonWebTokenLifetime, errorValue := time.ParseDuration(
		environmentValue("JSON_WEB_TOKEN_LIFETIME", "15m"),
	)
	if errorValue != nil {
		return nil, ErrInvalidConfiguration
	}

	applicationConfiguration := &Configuration{
		ApplicationPort:        environmentValue("APPLICATION_PORT", "3000"),
		ApplicationEnvironment: environmentValue("APPLICATION_ENVIRONMENT", "development"),
		PostgreSQLDataSource:   environmentValue("POSTGRESQL_DATA_SOURCE", "host=localhost port=5432 user=afraniocaires password=postgres dbname=ecommerce sslmode=disable"),
		JSONWebTokenSecret:     environmentValue("JSON_WEB_TOKEN_SECRET", jsonWebTokenSecretFallback),
		JSONWebTokenIssuer:     environmentValue("JSON_WEB_TOKEN_ISSUER", "afranio"),
		JSONWebTokenLifetime:   jsonWebTokenLifetime,
	}

	if applicationConfiguration.JSONWebTokenSecret == "" {
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
