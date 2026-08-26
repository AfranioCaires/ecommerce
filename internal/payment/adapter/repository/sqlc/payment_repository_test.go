package paymentrepository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
	databasequeries "github.com/afraniocaires/ecommerce/internal/payment/platform/database/sqlc"
)

func TestPaymentRepositoryInsertsPayment(t *testing.T) {
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	repository := NewPaymentRepository(databasequeries.New(databaseConnection))
	createdAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	payment, _ := domain.NewPayment(
		"payment-1",
		"order-1",
		20000,
		domain.PaymentStatusApproved,
		createdAt,
	)

	mock.ExpectExec("INSERT INTO payments").
		WithArgs("payment-1", "order-1", int64(20000), "APPROVED", createdAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if errorValue := repository.Save(context.Background(), payment); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}
