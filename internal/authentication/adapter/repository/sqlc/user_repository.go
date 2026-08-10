package authenticationrepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

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
			Name:         user.Name,
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

func (repository *UserRepository) FindAll(
	applicationContext context.Context,
) ([]*domain.User, error) {
	customers, errorValue := repository.queries.ListCustomers(applicationContext)
	if errorValue != nil {
		return nil, errorValue
	}

	users := make([]*domain.User, 0, len(customers))
	for _, customer := range customers {
		user, errorValue := toUser(
			customer.ID,
			customer.Name,
			customer.Email,
			customer.PasswordHash,
			customer.Roles,
			customer.CreatedAt,
		)
		if errorValue != nil {
			return nil, errorValue
		}
		users = append(users, user)
	}
	return users, nil
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

	return toUser(customer.ID, customer.Name, customer.Email, customer.PasswordHash, customer.Roles, customer.CreatedAt)
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

	return toUser(customer.ID, customer.Name, customer.Email, customer.PasswordHash, customer.Roles, customer.CreatedAt)
}

func roleValues(roles []domain.Role) string {
	values := make([]string, len(roles))
	for index, role := range roles {
		values[index] = string(role)
	}

	return strings.Join(values, ",")
}

func toUser(
	userID string,
	name string,
	email string,
	passwordHash string,
	roleValues string,
	createdAt time.Time,
) (*domain.User, error) {
	roleParts := strings.Split(roleValues, ",")
	roles := make([]domain.Role, 0, len(roleParts))
	for _, rolePart := range roleParts {
		if rolePart != "" {
			roles = append(roles, domain.Role(rolePart))
		}
	}

	return domain.NewUser(
		userID,
		name,
		email,
		passwordHash,
		roles,
		createdAt,
	)
}
