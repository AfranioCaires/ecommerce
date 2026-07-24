package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	postgresmigration "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type MigrationDirection string

const (
	MigrationDirectionUp   MigrationDirection = "up"
	MigrationDirectionDown MigrationDirection = "down"
)

var ErrInvalidMigrationDirection = errors.New("the migration direction is invalid")

func ApplyMigrations(
	databaseConnection *sql.DB,
	direction MigrationDirection,
) error {
	if direction != MigrationDirectionUp && direction != MigrationDirectionDown {
		return ErrInvalidMigrationDirection
	}

	sourceDriver, errorValue := iofs.New(migrationFiles, "migrations")
	if errorValue != nil {
		return fmt.Errorf("create migration source: %w", errorValue)
	}
	defer sourceDriver.Close()

	databaseDriver, errorValue := postgresmigration.WithInstance(
		databaseConnection,
		&postgresmigration.Config{},
	)
	if errorValue != nil {
		return fmt.Errorf("create PostgreSQL migration driver: %w", errorValue)
	}

	migrator, errorValue := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"postgresql",
		databaseDriver,
	)
	if errorValue != nil {
		return fmt.Errorf("create migrator: %w", errorValue)
	}

	switch direction {
	case MigrationDirectionUp:
		errorValue = migrator.Up()
	case MigrationDirectionDown:
		errorValue = migrator.Down()
	}

	if errors.Is(errorValue, migrate.ErrNoChange) {
		return nil
	}
	if errorValue != nil {
		return fmt.Errorf("apply %s migrations: %w", direction, errorValue)
	}

	return nil
}
