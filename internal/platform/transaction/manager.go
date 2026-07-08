package transaction

import (
	stdcontext "context"

	"gorm.io/gorm"
)

type databaseContextKey string

const activeDatabaseConnectionKey databaseContextKey = "active_database_connection"

type Manager struct {
	databaseConnection *gorm.DB
}

func NewManager(databaseConnection *gorm.DB) *Manager {
	return &Manager{databaseConnection: databaseConnection}
}

func (manager *Manager) Execute(
	context stdcontext.Context,
	operation func(transactionContext stdcontext.Context) error,
) error {
	return manager.databaseConnection.
		WithContext(context).
		Transaction(func(transactionConnection *gorm.DB) error {
			transactionContext := contextWithDatabaseConnection(
				context,
				transactionConnection,
			)

			return operation(transactionContext)
		})
}

func DatabaseConnection(
	context stdcontext.Context,
	fallbackDatabaseConnection *gorm.DB,
) *gorm.DB {
	databaseConnection, available := context.Value(
		activeDatabaseConnectionKey,
	).(*gorm.DB)
	if available {
		return databaseConnection.WithContext(context)
	}

	return fallbackDatabaseConnection.WithContext(context)
}

func contextWithDatabaseConnection(
	context stdcontext.Context,
	databaseConnection *gorm.DB,
) stdcontext.Context {
	return stdcontext.WithValue(
		context,
		activeDatabaseConnectionKey,
		databaseConnection,
	)
}
