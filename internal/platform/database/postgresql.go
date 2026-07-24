package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const connectionTimeout = 5 * time.Second

func NewPostgreSQLConnection(dataSource string) (*sql.DB, error) {
	databaseConnection, errorValue := sql.Open("pgx", dataSource)
	if errorValue != nil {
		return nil, errorValue
	}

	databaseConnection.SetMaxOpenConns(25)
	databaseConnection.SetMaxIdleConns(5)
	databaseConnection.SetConnMaxLifetime(30 * time.Minute)

	connectionContext, cancel := context.WithTimeout(
		context.Background(),
		connectionTimeout,
	)
	defer cancel()

	if errorValue := databaseConnection.PingContext(connectionContext); errorValue != nil {
		_ = databaseConnection.Close()
		return nil, errorValue
	}

	return databaseConnection, nil
}
