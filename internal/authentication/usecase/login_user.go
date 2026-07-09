package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type LoginUserInput struct {
	Email    string
	Password string
}

type LoginUserOutput struct {
	AccessToken string
}

type LoginUserUseCase struct {
	userRepository     UserRepository
	passwordHasher     PasswordHasher
	accessTokenService AccessTokenService
	currentTime        func() time.Time
}

func NewLoginUserUseCase(
	userRepository UserRepository,
	passwordHasher PasswordHasher,
	accessTokenService AccessTokenService,
	currentTime func() time.Time,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepository:     userRepository,
		passwordHasher:     passwordHasher,
		accessTokenService: accessTokenService,
		currentTime:        currentTime,
	}
}

func (useCase *LoginUserUseCase) Execute(
	context context.Context,
	input LoginUserInput,
) (*LoginUserOutput, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))

	user, errorValue := useCase.userRepository.FindByEmail(
		context,
		normalizedEmail,
	)
	if errorValue != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if errorValue := useCase.passwordHasher.Compare(
		user.PasswordHash,
		input.Password,
	); errorValue != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, errorValue := useCase.accessTokenService.Generate(
		user.ID,
		user.Roles,
		useCase.currentTime(),
	)
	if errorValue != nil {
		return nil, errorValue
	}

	return &LoginUserOutput{AccessToken: accessToken}, nil
}
