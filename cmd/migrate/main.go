package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/afraniocaires/ecommerce/internal/platform/configuration"
	"github.com/afraniocaires/ecommerce/internal/platform/database"
)

func main() {
	directionValue := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	applicationConfiguration, errorValue := configuration.Load()
	if errorValue != nil {
		slog.Error("The application configuration could not be loaded.", "error", errorValue)
		os.Exit(1)
	}

	databaseConnection, errorValue := database.NewPostgreSQLConnection(
		applicationConfiguration.PostgreSQLDataSource,
	)
	if errorValue != nil {
		slog.Error("The PostgreSQL connection could not be created.", "error", errorValue)
		os.Exit(1)
	}
	defer databaseConnection.Close()

	errorValue = database.ApplyMigrations(
		databaseConnection,
		database.MigrationDirection(*directionValue),
	)
	if errorValue != nil {
		slog.Error("The database migrations could not be applied.", "error", errorValue)
		os.Exit(1)
	}

	slog.Info("The database migrations were applied.", "direction", *directionValue)
}
