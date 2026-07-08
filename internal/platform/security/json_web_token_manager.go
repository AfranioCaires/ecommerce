package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

var ErrInvalidAccessToken = errors.New("the access token is invalid.")

type AccessTokenClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type JSONWebTokenManager struct {
	secret   []byte
	issuer   string
	lifetime time.Duration
}

func NewJSONWebTokenManager(
	secret string,
	issuer string,
	lifetime time.Duration,
) *JSONWebTokenManager {
	return &JSONWebTokenManager{
		secret:   []byte(secret),
		issuer:   issuer,
		lifetime: lifetime,
	}
}

func (tokenManager *JSONWebTokenManager) Generate(
	userID string,
	roles []domain.Role,
	issuedAt time.Time,
) (string, error) {
	roleValues := make([]string, len(roles))

	for index, role := range roles {
		roleValues[index] = string(role)
	}

	accessTokenClaims := AccessTokenClaims{
		Roles: roleValues,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    tokenManager.issuer,
			IssuedAt:  jwt.NewNumericDate(issuedAt.UTC()),
			ExpiresAt: jwt.NewNumericDate(issuedAt.UTC().Add(tokenManager.lifetime)),
		},
	}

	accessToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		accessTokenClaims,
	)

	return accessToken.SignedString(tokenManager.secret)
}

func (tokenManager *JSONWebTokenManager) Parse(
	accessTokenValue string,
) (*AccessTokenClaims, error) {
	accessTokenClaims := &AccessTokenClaims{}

	accessToken, errorValue := jwt.ParseWithClaims(
		accessTokenValue,
		accessTokenClaims,
		func(accessToken *jwt.Token) (any, error) {
			return tokenManager.secret, nil
		},
		jwt.WithIssuer(tokenManager.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if errorValue != nil || !accessToken.Valid {
		return nil, ErrInvalidAccessToken
	}

	return accessTokenClaims, nil
}
