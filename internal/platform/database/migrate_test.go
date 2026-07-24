package database

import (
	"errors"
	"strings"
	"testing"
)

func TestInitialMigrationDefinesRequiredSchema(t *testing.T) {
	migration, errorValue := migrationFiles.ReadFile(
		"migrations/000001_initial_schema.up.sql",
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}

	schema := string(migration)
	requiredFragments := []string{
		"CREATE TABLE customers",
		"password_hash TEXT NOT NULL",
		"customer_id TEXT NOT NULL REFERENCES customers(id)",
		"order_id TEXT NOT NULL REFERENCES orders(id)",
		"product_id TEXT NOT NULL REFERENCES products(id)",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Errorf("initial migration does not contain %q", fragment)
		}
	}
}

func TestInitialMigrationIsReversible(t *testing.T) {
	migration, errorValue := migrationFiles.ReadFile(
		"migrations/000001_initial_schema.down.sql",
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}

	for _, table := range []string{"payments", "order_items", "orders", "stocks", "products", "customers"} {
		if !strings.Contains(string(migration), "DROP TABLE IF EXISTS "+table) {
			t.Errorf("down migration does not drop %s", table)
		}
	}
}

func TestApplyMigrationsRejectsInvalidDirection(t *testing.T) {
	errorValue := ApplyMigrations(nil, MigrationDirection("sideways"))
	if !errors.Is(errorValue, ErrInvalidMigrationDirection) {
		t.Fatalf("ApplyMigrations() error = %v", errorValue)
	}
}
