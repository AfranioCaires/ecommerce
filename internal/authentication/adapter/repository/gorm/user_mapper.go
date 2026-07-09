package authenticationrepository

import (
	"strings"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

func toUserModel(user *domain.User) UserModel {
	roleValues := make([]string, len(user.Roles))

	for index, role := range user.Roles {
		roleValues[index] = string(role)
	}

	return UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Roles:        strings.Join(roleValues, ","),
		CreatedAt:    user.CreatedAt,
	}
}

func toUserEntity(userModel *UserModel) (*domain.User, error) {
	roleParts := strings.Split(userModel.Roles, ",")
	roles := make([]domain.Role, 0, len(roleParts))

	for _, rolePart := range roleParts {
		if rolePart != "" {
			roles = append(roles, domain.Role(rolePart))
		}
	}

	return domain.NewUser(
		userModel.ID,
		userModel.Email,
		userModel.PasswordHash,
		roles,
		userModel.CreatedAt,
	)
}
