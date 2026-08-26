package paymentrepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	databasequeries "github.com/afraniocaires/ecommerce/internal/payment/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/payment/platform/transaction"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type MessageRepository struct{ queries *databasequeries.Queries }

func NewMessageRepository(queries *databasequeries.Queries) *MessageRepository {
	return &MessageRepository{queries: queries}
}

func (repository *MessageRepository) Save(applicationContext context.Context, message *outbox.Message) error {
	return transaction.Queries(applicationContext, repository.queries).CreatePaymentOutboxMessage(applicationContext, databasequeries.CreatePaymentOutboxMessageParams{
		ID: message.ID, MessageType: message.MessageType, RoutingKey: message.RoutingKey, Payload: message.Payload,
		Attempts: int32(message.Attempts), NextAttemptAt: message.NextAttemptAt, PublishedAt: paymentNullableTime(message.PublishedAt), LastError: message.LastError, CreatedAt: message.CreatedAt,
	})
}
func (repository *MessageRepository) Pending(applicationContext context.Context, now time.Time, limit int) ([]*outbox.Message, error) {
	rows, errorValue := transaction.Queries(applicationContext, repository.queries).ListPendingPaymentOutboxMessages(applicationContext, databasequeries.ListPendingPaymentOutboxMessagesParams{NextAttemptAt: now.UTC(), Limit: int32(limit)})
	if errorValue != nil {
		return nil, errorValue
	}
	messages := make([]*outbox.Message, 0, len(rows))
	for _, row := range rows {
		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			value := row.PublishedAt.Time
			publishedAt = &value
		}
		messages = append(messages, &outbox.Message{ID: row.ID, MessageType: row.MessageType, RoutingKey: row.RoutingKey, Payload: append([]byte(nil), row.Payload...), Attempts: int(row.Attempts), NextAttemptAt: row.NextAttemptAt, PublishedAt: publishedAt, LastError: row.LastError, CreatedAt: row.CreatedAt})
	}
	return messages, nil
}
func (repository *MessageRepository) MarkPublished(applicationContext context.Context, id string, at time.Time) error {
	return transaction.Queries(applicationContext, repository.queries).MarkPaymentOutboxPublished(applicationContext, databasequeries.MarkPaymentOutboxPublishedParams{ID: id, PublishedAt: sql.NullTime{Time: at.UTC(), Valid: true}})
}
func (repository *MessageRepository) MarkFailed(applicationContext context.Context, id string, next time.Time, message string) error {
	if len(message) > 500 {
		message = message[:500]
	}
	return transaction.Queries(applicationContext, repository.queries).MarkPaymentOutboxFailed(applicationContext, databasequeries.MarkPaymentOutboxFailedParams{ID: id, NextAttemptAt: next.UTC(), LastError: message})
}
func (repository *MessageRepository) TrySave(applicationContext context.Context, id string, at time.Time) (bool, error) {
	errorValue := transaction.Queries(applicationContext, repository.queries).CreatePaymentInboxMessage(applicationContext, databasequeries.CreatePaymentInboxMessageParams{MessageID: id, ProcessedAt: at.UTC()})
	var databaseError *pgconn.PgError
	if errors.As(errorValue, &databaseError) && databaseError.Code == "23505" {
		return false, nil
	}
	return errorValue == nil, errorValue
}
func paymentNullableTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}
