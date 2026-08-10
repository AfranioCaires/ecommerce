package authenticationrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

func TestUserRepository(t *testing.T) {
	t.Run("it should save and find a user by normalized email", func(t *testing.T) {
		repository := NewUserRepository()
		user, _ := domain.NewUser("user-1", "Customer", "customer@example.com", "hash", nil, time.Now())
		if errorValue := repository.Save(context.Background(), user); errorValue != nil {
			t.Fatal(errorValue)
		}
		storedUser, errorValue := repository.FindByEmail(context.Background(), " CUSTOMER@EXAMPLE.COM ")
		if errorValue != nil || storedUser.ID != user.ID {
			t.Fatalf("expected stored user, received %#v and %v", storedUser, errorValue)
		}
	})

	t.Run("it should reject a duplicate email", func(t *testing.T) {
		repository := NewUserRepository()
		firstUser, _ := domain.NewUser("user-1", "Customer", "customer@example.com", "hash", nil, time.Now())
		secondUser, _ := domain.NewUser("user-2", "Other", "customer@example.com", "hash", nil, time.Now())
		repository.Save(context.Background(), firstUser)
		errorValue := repository.Save(context.Background(), secondUser)
		if !errors.Is(errorValue, domain.ErrEmailAlreadyUsed) {
			t.Fatalf("expected duplicate email, received %v", errorValue)
		}
	})

	t.Run("it should find a user by ID and return independent role slices", func(t *testing.T) {
		repository := NewUserRepository()
		user, _ := domain.NewUser("user-1", "Customer", "customer@example.com", "hash", []domain.Role{domain.RoleCustomer}, time.Now())
		if errorValue := repository.Save(context.Background(), user); errorValue != nil {
			t.Fatal(errorValue)
		}

		storedUser, errorValue := repository.FindByID(context.Background(), user.ID)
		if errorValue != nil || storedUser.ID != user.ID {
			t.Fatalf("expected stored user, received %#v and %v", storedUser, errorValue)
		}
		storedUser.Roles[0] = domain.RoleAdministrator
		storedAgain, _ := repository.FindByEmail(context.Background(), user.Email)
		if storedAgain.Roles[0] != domain.RoleCustomer {
			t.Fatal("expected repository role slice to be isolated")
		}
	})

	t.Run("it should return user not found for missing identifiers", func(t *testing.T) {
		repository := NewUserRepository()
		if _, errorValue := repository.FindByEmail(context.Background(), "missing@example.com"); !errors.Is(errorValue, domain.ErrUserNotFound) {
			t.Fatalf("expected user not found by email, received %v", errorValue)
		}
		if _, errorValue := repository.FindByID(context.Background(), "missing"); !errors.Is(errorValue, domain.ErrUserNotFound) {
			t.Fatalf("expected user not found by ID, received %v", errorValue)
		}
	})
}
