package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterUserUseCase struct {
	userRepository UserRepository
	passwordHasher PasswordHasher
	currentTime    func() time.Time
}

func NewRegisterUserUseCase(
	userRepository UserRepository,
	passwordHasher PasswordHasher,
	currentTime func() time.Time,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		currentTime:    currentTime,
	}
}

func (useCase *RegisterUserUseCase) Execute(
	context context.Context,
	input RegisterUserInput,
) (*domain.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))
	normalizedName := strings.TrimSpace(input.Name)
	if normalizedName == "" {
		normalizedName = strings.SplitN(normalizedEmail, "@", 2)[0]
	}

	existingUser, errorValue := useCase.userRepository.FindByEmail(
		context,
		normalizedEmail,
	)
	if errorValue == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyUsed
	}
	if errorValue != nil && !errors.Is(errorValue, domain.ErrUserNotFound) {
		return nil, errorValue
	}

	passwordHash, errorValue := useCase.passwordHasher.Hash(input.Password)
	if errorValue != nil {
		return nil, errorValue
	}

	user, errorValue := domain.NewUser(
		uuid.NewString(),
		normalizedName,
		normalizedEmail,
		passwordHash,
		[]domain.Role{domain.RoleCustomer},
		useCase.currentTime(),
	)
	if errorValue != nil {
		return nil, errorValue
	}

	if errorValue := useCase.userRepository.Save(context, user); errorValue != nil {
		return nil, errorValue
	}

	return user, nil
}
