package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	authenticationrepository "github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

func TestAuthenticationUseCases(t *testing.T) {
	currentTime := func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }
	userRepository := authenticationrepository.NewUserRepository()
	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.MinCost)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	registerUserUseCase := usecase.NewRegisterUserUseCase(userRepository, passwordHasher, currentTime)
	loginUserUseCase := usecase.NewLoginUserUseCase(userRepository, passwordHasher, accessTokenManager, currentTime)

	t.Run("it should register and authenticate a customer", func(t *testing.T) {
		user, errorValue := registerUserUseCase.Execute(context.Background(), usecase.RegisterUserInput{Email: " Customer@Example.com ", Password: "password"})
		if errorValue != nil || user.Email != "customer@example.com" || !user.HasRole(domain.RoleCustomer) {
			t.Fatalf("unexpected registration: %#v, %v", user, errorValue)
		}
		output, errorValue := loginUserUseCase.Execute(context.Background(), usecase.LoginUserInput{Email: user.Email, Password: "password"})
		if errorValue != nil || output.AccessToken == "" {
			t.Fatalf("unexpected login: %#v, %v", output, errorValue)
		}
	})

	t.Run("it should reject a duplicate email", func(t *testing.T) {
		_, errorValue := registerUserUseCase.Execute(context.Background(), usecase.RegisterUserInput{Email: "customer@example.com", Password: "password"})
		if !errors.Is(errorValue, domain.ErrEmailAlreadyUsed) {
			t.Fatalf("expected duplicate email, received %v", errorValue)
		}
	})

	t.Run("it should reject invalid credentials", func(t *testing.T) {
		_, errorValue := loginUserUseCase.Execute(context.Background(), usecase.LoginUserInput{Email: "customer@example.com", Password: "wrong"})
		if !errors.Is(errorValue, domain.ErrInvalidCredentials) {
			t.Fatalf("expected invalid credentials, received %v", errorValue)
		}
	})
}
