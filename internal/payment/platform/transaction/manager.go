package transaction

import (
	"context"
	"database/sql"
	"fmt"

	databasequeries "github.com/afraniocaires/ecommerce/internal/payment/platform/database/sqlc"
)

type queriesContextKey struct{}
type Manager struct{ databaseConnection *sql.DB }

func NewManager(databaseConnection *sql.DB) *Manager {
	return &Manager{databaseConnection: databaseConnection}
}

func (manager *Manager) Execute(applicationContext context.Context, operation func(context.Context) error) error {
	databaseTransaction, errorValue := manager.databaseConnection.BeginTx(applicationContext, nil)
	if errorValue != nil {
		return fmt.Errorf("begin payment transaction: %w", errorValue)
	}
	transactionContext := context.WithValue(applicationContext, queriesContextKey{}, databasequeries.New(databaseTransaction))
	if errorValue := operation(transactionContext); errorValue != nil {
		if rollbackError := databaseTransaction.Rollback(); rollbackError != nil {
			return fmt.Errorf("rollback payment transaction after %v: %w", errorValue, rollbackError)
		}
		return errorValue
	}
	if errorValue := databaseTransaction.Commit(); errorValue != nil {
		return fmt.Errorf("commit payment transaction: %w", errorValue)
	}
	return nil
}

func Queries(applicationContext context.Context, fallback *databasequeries.Queries) *databasequeries.Queries {
	if queries, available := applicationContext.Value(queriesContextKey{}).(*databasequeries.Queries); available {
		return queries
	}
	return fallback
}
