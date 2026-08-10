package usecase

import (
	"context"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type ListUsersUseCase struct{ userRepository UserRepository }

func NewListUsersUseCase(userRepository UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{userRepository: userRepository}
}

func (useCase *ListUsersUseCase) Execute(context context.Context) ([]*domain.User, error) {
	return useCase.userRepository.FindAll(context)
}
