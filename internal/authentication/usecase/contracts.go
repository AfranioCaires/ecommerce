package usecase

import (
	"context"
	"time"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type UserRepository interface {
	Save(context context.Context, user *domain.User) error
	FindByEmail(context context.Context, email string) (*domain.User, error)
	FindByID(context context.Context, userID string) (*domain.User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(passwordHash string, password string) error
}

type AccessTokenService interface {
	Generate(
		userID string,
		roles []domain.Role,
		issuedAt time.Time,
	) (string, error)
}
