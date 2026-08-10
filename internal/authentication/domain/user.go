package domain

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var (
	ErrEmptyUserID        = errors.New("the user ID must not be empty.")
	ErrEmptyUserName      = errors.New("the user name must not be empty.")
	ErrEmptyEmail         = errors.New("the email must not be empty.")
	ErrEmptyPasswordHash  = errors.New("the password hash must not be empty.")
	ErrEmailAlreadyUsed   = errors.New("the email is already registered.")
	ErrUserNotFound       = errors.New("the user was not found.")
	ErrInvalidCredentials = errors.New("the credentials are invalid.")
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	Roles        []Role
	CreatedAt    time.Time
}

func NewUser(
	userID string,
	name string,
	email string,
	passwordHash string,
	roles []Role,
	createdAt time.Time,
) (*User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrEmptyUserID
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return nil, ErrEmptyEmail
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return nil, ErrEmptyUserName
	}

	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	if len(roles) == 0 {
		roles = []Role{RoleCustomer}
	}

	for _, role := range roles {
		if !role.IsValid() {
			return nil, ErrInvalidRole
		}
	}

	return &User{
		ID:           userID,
		Name:         normalizedName,
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
		Roles:        roles,
		CreatedAt:    createdAt.UTC(),
	}, nil
}

func (user *User) HasRole(requiredRole Role) bool {
	return slices.Contains(user.Roles, requiredRole)
}
