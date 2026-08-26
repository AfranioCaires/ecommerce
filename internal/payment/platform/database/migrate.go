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

func ApplyMigrations(databaseConnection *sql.DB) error {
	source, errorValue := iofs.New(migrationFiles, "migrations")
	if errorValue != nil {
		return errorValue
	}
	defer source.Close()
	driver, errorValue := postgresmigration.WithInstance(databaseConnection, &postgresmigration.Config{})
	if errorValue != nil {
		return errorValue
	}
	migrator, errorValue := migrate.NewWithInstance("iofs", source, "postgresql", driver)
	if errorValue != nil {
		return errorValue
	}
	if errorValue = migrator.Up(); errors.Is(errorValue, migrate.ErrNoChange) {
		return nil
	}
	if errorValue != nil {
		return fmt.Errorf("apply payment migrations: %w", errorValue)
	}
	return nil
}
