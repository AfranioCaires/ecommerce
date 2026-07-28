package authenticationrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
)

func newUserRepository(t *testing.T) (*UserRepository, sqlmock.Sqlmock) {
	t.Helper()
	databaseConnection, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = databaseConnection.Close() })

	return NewUserRepository(databasequeries.New(databaseConnection)), mock
}

func TestUserRepositorySavesCustomerPasswordHash(t *testing.T) {
	repository, mock := newUserRepository(t)
	createdAt := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	user, _ := domain.NewUser(
		"customer-1",
		"customer@example.com",
		"password-hash",
		[]domain.Role{domain.RoleCustomer, domain.RoleAdministrator},
		createdAt,
	)

	mock.ExpectExec("INSERT INTO customers").
		WithArgs(
			"customer-1",
			"customer@example.com",
			"password-hash",
			"CUSTOMER,ADMIN",
			createdAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if errorValue := repository.Save(context.Background(), user); errorValue != nil {
		t.Fatalf("Save() error = %v", errorValue)
	}
	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}

func TestUserRepositoryMapsUniqueEmailViolation(t *testing.T) {
	repository, mock := newUserRepository(t)
	user, _ := domain.NewUser(
		"customer-1",
		"customer@example.com",
		"password-hash",
		[]domain.Role{domain.RoleCustomer},
		time.Now(),
	)

	mock.ExpectExec("INSERT INTO customers").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	if errorValue := repository.Save(context.Background(), user); !errors.Is(errorValue, domain.ErrEmailAlreadyUsed) {
		t.Fatalf("Save() error = %v", errorValue)
	}
}

func TestUserRepositoryFindsCustomers(t *testing.T) {
	repository, mock := newUserRepository(t)
	createdAt := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	columns := []string{"id", "email", "password_hash", "roles", "created_at"}

	mock.ExpectQuery("FROM customers").
		WithArgs("customer@example.com").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(
			"customer-1",
			"customer@example.com",
			"password-hash",
			"CUSTOMER,SUPPORT",
			createdAt,
		))
	user, errorValue := repository.FindByEmail(context.Background(), "customer@example.com")
	if errorValue != nil || user.ID != "customer-1" || user.PasswordHash != "password-hash" || len(user.Roles) != 2 {
		t.Fatalf("FindByEmail() = %#v, %v", user, errorValue)
	}

	mock.ExpectQuery("FROM customers").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)
	if _, errorValue := repository.FindByID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrUserNotFound) {
		t.Fatalf("FindByID() error = %v", errorValue)
	}

	if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
		t.Fatal(errorValue)
	}
}
