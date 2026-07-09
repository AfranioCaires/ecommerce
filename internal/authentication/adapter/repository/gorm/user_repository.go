package authenticationrepository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
)

type UserRepository struct {
	databaseConnection *gorm.DB
}

func NewUserRepository(databaseConnection *gorm.DB) *UserRepository {
	return &UserRepository{databaseConnection: databaseConnection}
}

var _ usecase.UserRepository = (*UserRepository)(nil)

func (repository *UserRepository) Save(
	context context.Context,
	user *domain.User,
) error {
	userModel := toUserModel(user)

	errorValue := repository.databaseConnection.
		WithContext(context).
		Create(&userModel).
		Error

	if errorValue != nil {
		return errorValue
	}

	return nil
}

func (repository *UserRepository) FindByEmail(
	context context.Context,
	email string,
) (*domain.User, error) {
	var userModel UserModel

	errorValue := repository.databaseConnection.
		WithContext(context).
		First(&userModel, "email = ?", email).
		Error

	if errors.Is(errorValue, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toUserEntity(&userModel)
}

func (repository *UserRepository) FindByID(
	context context.Context,
	userID string,
) (*domain.User, error) {
	var userModel UserModel

	errorValue := repository.databaseConnection.
		WithContext(context).
		First(&userModel, "id = ?", userID).
		Error

	if errors.Is(errorValue, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toUserEntity(&userModel)
}
