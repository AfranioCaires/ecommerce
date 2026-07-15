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

type userRepositoryStub struct {
	findByEmail func(string) (*domain.User, error)
	save        func(*domain.User) error
}

func (stub userRepositoryStub) Save(_ context.Context, user *domain.User) error {
	if stub.save == nil {
		return nil
	}
	return stub.save(user)
}

func (stub userRepositoryStub) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return stub.findByEmail(email)
}

func (stub userRepositoryStub) FindByID(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

type passwordHasherStub struct {
	hash    func(string) (string, error)
	compare func(string, string) error
}

func (stub passwordHasherStub) Hash(password string) (string, error) {
	return stub.hash(password)
}

func (stub passwordHasherStub) Compare(hash string, password string) error {
	return stub.compare(hash, password)
}

type accessTokenServiceStub struct {
	generate func(string, []domain.Role, time.Time) (string, error)
}

func (stub accessTokenServiceStub) Generate(userID string, roles []domain.Role, issuedAt time.Time) (string, error) {
	return stub.generate(userID, roles, issuedAt)
}

func TestRegisterUserErrors(t *testing.T) {
	expectedError := errors.New("dependency failed")
	currentTime := func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }

	t.Run("it should return an unexpected lookup error", func(t *testing.T) {
		useCase := usecase.NewRegisterUserUseCase(
			userRepositoryStub{findByEmail: func(string) (*domain.User, error) { return nil, expectedError }},
			passwordHasherStub{}, currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.RegisterUserInput{Email: "a@example.com", Password: "password"}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected dependency error, received %v", errorValue)
		}
	})

	t.Run("it should return a hash error", func(t *testing.T) {
		useCase := usecase.NewRegisterUserUseCase(
			userRepositoryStub{findByEmail: func(string) (*domain.User, error) { return nil, domain.ErrUserNotFound }},
			passwordHasherStub{hash: func(string) (string, error) { return "", expectedError }}, currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.RegisterUserInput{Email: "a@example.com", Password: "password"}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected hash error, received %v", errorValue)
		}
	})

	t.Run("it should return a domain validation error", func(t *testing.T) {
		useCase := usecase.NewRegisterUserUseCase(
			userRepositoryStub{findByEmail: func(string) (*domain.User, error) { return nil, domain.ErrUserNotFound }},
			passwordHasherStub{hash: func(string) (string, error) { return "hash", nil }}, currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.RegisterUserInput{Email: " ", Password: "password"}); !errors.Is(errorValue, domain.ErrEmptyEmail) {
			t.Fatalf("expected empty email, received %v", errorValue)
		}
	})

	t.Run("it should return a save error", func(t *testing.T) {
		useCase := usecase.NewRegisterUserUseCase(
			userRepositoryStub{
				findByEmail: func(string) (*domain.User, error) { return nil, domain.ErrUserNotFound },
				save:        func(*domain.User) error { return expectedError },
			},
			passwordHasherStub{hash: func(string) (string, error) { return "hash", nil }}, currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.RegisterUserInput{Email: "a@example.com", Password: "password"}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected save error, received %v", errorValue)
		}
	})
}

func TestLoginUserErrors(t *testing.T) {
	expectedError := errors.New("token failed")
	user, _ := domain.NewUser("user-1", "a@example.com", "hash", nil, time.Now())
	currentTime := func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }

	t.Run("it should hide lookup errors", func(t *testing.T) {
		useCase := usecase.NewLoginUserUseCase(
			userRepositoryStub{findByEmail: func(string) (*domain.User, error) { return nil, domain.ErrUserNotFound }},
			passwordHasherStub{}, accessTokenServiceStub{}, currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.LoginUserInput{Email: "missing@example.com"}); !errors.Is(errorValue, domain.ErrInvalidCredentials) {
			t.Fatalf("expected invalid credentials, received %v", errorValue)
		}
	})

	t.Run("it should return a token generation error", func(t *testing.T) {
		useCase := usecase.NewLoginUserUseCase(
			userRepositoryStub{findByEmail: func(email string) (*domain.User, error) {
				if email != "a@example.com" {
					t.Fatalf("expected normalized email, received %q", email)
				}
				return user, nil
			}},
			passwordHasherStub{compare: func(string, string) error { return nil }},
			accessTokenServiceStub{generate: func(string, []domain.Role, time.Time) (string, error) { return "", expectedError }},
			currentTime,
		)
		if _, errorValue := useCase.Execute(context.Background(), usecase.LoginUserInput{Email: " A@EXAMPLE.COM "}); !errors.Is(errorValue, expectedError) {
			t.Fatalf("expected token error, received %v", errorValue)
		}
	})
}
