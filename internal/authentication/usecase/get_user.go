package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type GetUserUseCase struct{ userRepository UserRepository }

func NewGetUserUseCase(userRepository UserRepository) *GetUserUseCase {
	return &GetUserUseCase{userRepository: userRepository}
}

func (useCase *GetUserUseCase) Execute(context context.Context, userID string) (*domain.User, error) {
	return useCase.userRepository.FindByID(context, userID)
}
