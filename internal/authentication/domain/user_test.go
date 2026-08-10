package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	t.Run("it should normalize email and assign the customer role", func(t *testing.T) {
		user, errorValue := NewUser("user-1", " Customer ", " CUSTOMER@EXAMPLE.COM ", "hash", nil, time.Now())
		if errorValue != nil || user.Name != "Customer" || user.Email != "customer@example.com" || !user.HasRole(RoleCustomer) {
			t.Fatalf("unexpected user: %#v, %v", user, errorValue)
		}
	})

	for description, input := range map[string]struct {
		userID       string
		name         string
		email        string
		passwordHash string
		roles        []Role
		expected     error
	}{
		"it should reject an empty user ID":       {"", "Customer", "a@example.com", "hash", nil, ErrEmptyUserID},
		"it should reject an empty name":          {"user-1", " ", "a@example.com", "hash", nil, ErrEmptyUserName},
		"it should reject an empty email":         {"user-1", "Customer", " ", "hash", nil, ErrEmptyEmail},
		"it should reject an empty password hash": {"user-1", "Customer", "a@example.com", "", nil, ErrEmptyPasswordHash},
		"it should reject an invalid role":        {"user-1", "Customer", "a@example.com", "hash", []Role{"INVALID"}, ErrInvalidRole},
	} {
		t.Run(description, func(t *testing.T) {
			_, errorValue := NewUser(input.userID, input.name, input.email, input.passwordHash, input.roles, time.Now())
			if !errors.Is(errorValue, input.expected) {
				t.Fatalf("expected %v, received %v", input.expected, errorValue)
			}
		})
	}
}
