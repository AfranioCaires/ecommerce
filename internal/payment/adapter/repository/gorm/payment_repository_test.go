package paymentrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/payment/domain"
)

func newMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDatabase, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	databaseConnection, errorValue := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDatabase}),
		&gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true},
	)
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return databaseConnection, mock
}

func TestPaymentModelTableName(t *testing.T) {
	if tableName := (PaymentModel{}).TableName(); tableName != "payments" {
		t.Fatalf("TableName() = %q", tableName)
	}
}

func TestPaymentRepositorySave(t *testing.T) {
	databaseConnection, mock := newMockDatabase(t)
	repository := NewPaymentRepository(databaseConnection)
	createdAt := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	payment, _ := domain.NewPayment("payment-1", "order-1", 1500, domain.PaymentStatusApproved, createdAt)

	mock.ExpectExec(`INSERT INTO "payments"`).
		WithArgs(payment.ID, payment.OrderID, payment.AmountCents, string(payment.Status), payment.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if errorValue := repository.Save(context.Background(), payment); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestPaymentRepositorySaveError(t *testing.T) {
	databaseConnection, mock := newMockDatabase(t)
	repository := NewPaymentRepository(databaseConnection)
	payment, _ := domain.NewPayment("payment-1", "order-1", 1500, domain.PaymentStatusDeclined, time.Now())
	expectedError := errors.New("insert failed")

	mock.ExpectExec(`INSERT INTO "payments"`).WillReturnError(expectedError)
	if errorValue := repository.Save(context.Background(), payment); !errors.Is(errorValue, expectedError) {
		t.Fatalf("Save() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}
