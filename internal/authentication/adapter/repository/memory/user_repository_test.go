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
		user, _ := domain.NewUser("user-1", "customer@example.com", "hash", nil, time.Now())
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
		firstUser, _ := domain.NewUser("user-1", "customer@example.com", "hash", nil, time.Now())
		secondUser, _ := domain.NewUser("user-2", "customer@example.com", "hash", nil, time.Now())
		repository.Save(context.Background(), firstUser)
		errorValue := repository.Save(context.Background(), secondUser)
		if !errors.Is(errorValue, domain.ErrEmailAlreadyUsed) {
			t.Fatalf("expected duplicate email, received %v", errorValue)
		}
	})
}
