package inbox

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type Repository interface {
	TrySave(context.Context, string, time.Time) (bool, error)
}

type SQLCRepository struct{ queries *databasequeries.Queries }

func NewSQLCRepository(queries *databasequeries.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (repository *SQLCRepository) TrySave(applicationContext context.Context, messageID string, processedAt time.Time) (bool, error) {
	errorValue := transaction.Queries(applicationContext, repository.queries).CreateInboxMessage(applicationContext, databasequeries.CreateInboxMessageParams{MessageID: messageID, ProcessedAt: processedAt.UTC()})
	var databaseError *pgconn.PgError
	if errors.As(errorValue, &databaseError) && databaseError.Code == "23505" {
		return false, nil
	}
	return errorValue == nil, errorValue
}
