package authenticationrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
)

type UserRepository struct {
	queries *databasequeries.Queries
}

func NewUserRepository(queries *databasequeries.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

var _ usecase.UserRepository = (*UserRepository)(nil)

func (repository *UserRepository) Save(
	applicationContext context.Context,
	user *domain.User,
) error {
	errorValue := repository.queries.CreateCustomer(
		applicationContext,
		databasequeries.CreateCustomerParams{
			ID:           user.ID,
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
			Roles:        roleValues(user.Roles),
			CreatedAt:    user.CreatedAt,
		},
	)

	var databaseError *pgconn.PgError
	if errors.As(errorValue, &databaseError) && databaseError.Code == "23505" {
		return domain.ErrEmailAlreadyUsed
	}

	return errorValue
}

func (repository *UserRepository) FindByEmail(
	applicationContext context.Context,
	email string,
) (*domain.User, error) {
	customer, errorValue := repository.queries.GetCustomerByEmail(
		applicationContext,
		email,
	)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toUser(customer)
}

func (repository *UserRepository) FindByID(
	applicationContext context.Context,
	userID string,
) (*domain.User, error) {
	customer, errorValue := repository.queries.GetCustomerByID(
		applicationContext,
		userID,
	)
	if errors.Is(errorValue, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if errorValue != nil {
		return nil, errorValue
	}

	return toUser(customer)
}

func roleValues(roles []domain.Role) string {
	values := make([]string, len(roles))
	for index, role := range roles {
		values[index] = string(role)
	}

	return strings.Join(values, ",")
}

func toUser(customer databasequeries.Customer) (*domain.User, error) {
	roleParts := strings.Split(customer.Roles, ",")
	roles := make([]domain.Role, 0, len(roleParts))
	for _, rolePart := range roleParts {
		if rolePart != "" {
			roles = append(roles, domain.Role(rolePart))
		}
	}

	return domain.NewUser(
		customer.ID,
		customer.Email,
		customer.PasswordHash,
		roles,
		customer.CreatedAt,
	)
}
