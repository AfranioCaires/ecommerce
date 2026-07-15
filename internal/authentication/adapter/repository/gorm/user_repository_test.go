package authenticationrepository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

var errUserDatabase = errors.New("user database failed")

func newUserMockDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDatabase, mock, errorValue := sqlmock.New()
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	t.Cleanup(func() { _ = sqlDatabase.Close() })
	databaseConnection, errorValue := gorm.Open(postgres.New(postgres.Config{Conn: sqlDatabase}), &gorm.Config{DisableAutomaticPing: true})
	if errorValue != nil {
		t.Fatal(errorValue)
	}
	return databaseConnection, mock
}

func userRows(createdAt time.Time, roles string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "email", "password_hash", "roles", "created_at"}).
		AddRow("user-1", "user@example.com", "hash", roles, createdAt)
}

func TestUserModelAndMappers(t *testing.T) {
	createdAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	user, _ := domain.NewUser("user-1", "USER@EXAMPLE.COM", "hash", []domain.Role{domain.RoleCustomer, domain.RoleAdministrator}, createdAt)
	model := toUserModel(user)
	if (UserModel{}).TableName() != "users" || model.ID != user.ID || model.Email != user.Email || model.PasswordHash != user.PasswordHash || model.Roles != "CUSTOMER,ADMIN" || !model.CreatedAt.Equal(createdAt) {
		t.Fatalf("toUserModel() = %#v", model)
	}

	entity, errorValue := toUserEntity(&UserModel{ID: "user-1", Email: "USER@EXAMPLE.COM", PasswordHash: "hash", Roles: ",CUSTOMER,", CreatedAt: createdAt})
	if errorValue != nil || entity.Email != "user@example.com" || len(entity.Roles) != 1 || entity.Roles[0] != domain.RoleCustomer {
		t.Fatalf("toUserEntity() = %#v, %v", entity, errorValue)
	}
	if entity, errorValue = toUserEntity(&UserModel{ID: "user-1", Email: "user@example.com", PasswordHash: "hash", Roles: "INVALID", CreatedAt: createdAt}); entity != nil || !errors.Is(errorValue, domain.ErrInvalidRole) {
		t.Fatalf("toUserEntity(invalid) = %#v, %v", entity, errorValue)
	}
}

func TestUserRepositorySave(t *testing.T) {
	createdAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	user, _ := domain.NewUser("user-1", "user@example.com", "hash", []domain.Role{domain.RoleCustomer}, createdAt)

	t.Run("success", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		if errorValue := NewUserRepository(database).Save(context.Background(), user); errorValue != nil {
			t.Fatalf("Save() error = %v", errorValue)
		}
		if errorValue := mock.ExpectationsWereMet(); errorValue != nil {
			t.Fatal(errorValue)
		}
	})

	t.Run("database error", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).WillReturnError(errUserDatabase)
		mock.ExpectRollback()
		if errorValue := NewUserRepository(database).Save(context.Background(), user); !errors.Is(errorValue, errUserDatabase) {
			t.Fatalf("Save() error = %v", errorValue)
		}
	})
}

func TestUserRepositoryFindByEmail(t *testing.T) {
	testUserFind(t, "email =", func(repository *UserRepository) (*domain.User, error) {
		return repository.FindByEmail(context.Background(), "user@example.com")
	})
}

func TestUserRepositoryFindByID(t *testing.T) {
	testUserFind(t, "id =", func(repository *UserRepository) (*domain.User, error) {
		return repository.FindByID(context.Background(), "user-1")
	})
}

func testUserFind(t *testing.T, condition string, find func(*UserRepository) (*domain.User, error)) {
	t.Helper()
	createdAt := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	query := `SELECT .* FROM "users" WHERE ` + regexp.QuoteMeta(condition)

	t.Run("not found", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "roles", "created_at"}))
		user, errorValue := find(NewUserRepository(database))
		if user != nil || !errors.Is(errorValue, domain.ErrUserNotFound) {
			t.Fatalf("find() = %#v, %v", user, errorValue)
		}
	})

	t.Run("database error", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectQuery(query).WillReturnError(errUserDatabase)
		user, errorValue := find(NewUserRepository(database))
		if user != nil || !errors.Is(errorValue, errUserDatabase) {
			t.Fatalf("find() = %#v, %v", user, errorValue)
		}
	})

	t.Run("invalid stored user", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectQuery(query).WillReturnRows(userRows(createdAt, "INVALID"))
		user, errorValue := find(NewUserRepository(database))
		if user != nil || !errors.Is(errorValue, domain.ErrInvalidRole) {
			t.Fatalf("find() = %#v, %v", user, errorValue)
		}
	})

	t.Run("success", func(t *testing.T) {
		database, mock := newUserMockDatabase(t)
		mock.ExpectQuery(query).WillReturnRows(userRows(createdAt, "CUSTOMER"))
		user, errorValue := find(NewUserRepository(database))
		if errorValue != nil || user.ID != "user-1" || user.Email != "user@example.com" {
			t.Fatalf("find() = %#v, %v", user, errorValue)
		}
	})
}
