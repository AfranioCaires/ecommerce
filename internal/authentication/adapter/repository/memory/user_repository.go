package authenticationrepository

import (
	"context"
	"strings"
	"sync"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

type UserRepository struct {
	mutex         sync.RWMutex
	usersByID     map[string]domain.User
	userIDByEmail map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		usersByID:     make(map[string]domain.User),
		userIDByEmail: make(map[string]string),
	}
}

func (repository *UserRepository) Save(applicationContext context.Context, user *domain.User) error {
	_ = applicationContext
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	if existingUserID, available := repository.userIDByEmail[normalizedEmail]; available && existingUserID != user.ID {
		return domain.ErrEmailAlreadyUsed
	}

	storedUser := copyUser(user)
	repository.usersByID[user.ID] = storedUser
	repository.userIDByEmail[normalizedEmail] = user.ID
	return nil
}

func (repository *UserRepository) FindByEmail(applicationContext context.Context, email string) (*domain.User, error) {
	_ = applicationContext
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	userID, available := repository.userIDByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !available {
		return nil, domain.ErrUserNotFound
	}
	user := repository.usersByID[userID]
	userCopy := copyUser(&user)
	return &userCopy, nil
}

func (repository *UserRepository) FindByID(applicationContext context.Context, userID string) (*domain.User, error) {
	_ = applicationContext
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	user, available := repository.usersByID[userID]
	if !available {
		return nil, domain.ErrUserNotFound
	}
	userCopy := copyUser(&user)
	return &userCopy, nil
}

func copyUser(user *domain.User) domain.User {
	userCopy := *user
	userCopy.Roles = append([]domain.Role(nil), user.Roles...)
	return userCopy
}
