package outbox

import (
	"context"
	"database/sql"
	"time"

	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type SQLCRepository struct{ queries *databasequeries.Queries }

func NewSQLCRepository(queries *databasequeries.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (repository *SQLCRepository) Save(applicationContext context.Context, message *Message) error {
	return transaction.Queries(applicationContext, repository.queries).CreateOutboxMessage(applicationContext, databasequeries.CreateOutboxMessageParams{
		ID: message.ID, MessageType: message.MessageType, RoutingKey: message.RoutingKey, Payload: message.Payload,
		Attempts: int32(message.Attempts), NextAttemptAt: message.NextAttemptAt, PublishedAt: nullableTime(message.PublishedAt), LastError: message.LastError, CreatedAt: message.CreatedAt,
	})
}

func (repository *SQLCRepository) Pending(applicationContext context.Context, now time.Time, limit int) ([]*Message, error) {
	rows, errorValue := transaction.Queries(applicationContext, repository.queries).ListPendingOutboxMessages(applicationContext, databasequeries.ListPendingOutboxMessagesParams{NextAttemptAt: now.UTC(), Limit: int32(limit)})
	if errorValue != nil {
		return nil, errorValue
	}
	messages := make([]*Message, 0, len(rows))
	for _, row := range rows {
		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			value := row.PublishedAt.Time
			publishedAt = &value
		}
		messages = append(messages, &Message{ID: row.ID, MessageType: row.MessageType, RoutingKey: row.RoutingKey, Payload: append([]byte(nil), row.Payload...), Attempts: int(row.Attempts), NextAttemptAt: row.NextAttemptAt, PublishedAt: publishedAt, LastError: row.LastError, CreatedAt: row.CreatedAt})
	}
	return messages, nil
}

func (repository *SQLCRepository) MarkPublished(applicationContext context.Context, id string, publishedAt time.Time) error {
	return transaction.Queries(applicationContext, repository.queries).MarkOutboxMessagePublished(applicationContext, databasequeries.MarkOutboxMessagePublishedParams{ID: id, PublishedAt: sql.NullTime{Time: publishedAt.UTC(), Valid: true}})
}

func (repository *SQLCRepository) MarkFailed(applicationContext context.Context, id string, nextAttemptAt time.Time, lastError string) error {
	if len(lastError) > 500 {
		lastError = lastError[:500]
	}
	return transaction.Queries(applicationContext, repository.queries).MarkOutboxMessageFailed(applicationContext, databasequeries.MarkOutboxMessageFailedParams{ID: id, NextAttemptAt: nextAttemptAt.UTC(), LastError: lastError})
}

func nullableTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}
