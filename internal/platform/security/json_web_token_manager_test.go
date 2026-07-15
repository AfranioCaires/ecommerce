package security

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
)

func TestJSONWebTokenManager(t *testing.T) {
	issuedAt := time.Now().Add(-time.Minute).Truncate(time.Second)
	manager := NewJSONWebTokenManager("secret", "issuer", time.Hour)
	tokenValue, err := manager.Generate(
		"user-1",
		[]domain.Role{domain.RoleCustomer, domain.RoleSupport},
		issuedAt,
	)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	claims, err := manager.Parse(tokenValue)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != "user-1" || claims.Issuer != "issuer" ||
		len(claims.Roles) != 2 || claims.Roles[0] != "CUSTOMER" || claims.Roles[1] != "SUPPORT" ||
		!claims.IssuedAt.Time.Equal(issuedAt.UTC()) ||
		!claims.ExpiresAt.Time.Equal(issuedAt.UTC().Add(time.Hour)) {
		t.Fatalf("Parse() claims = %#v", claims)
	}

	tests := []struct {
		name  string
		value func(t *testing.T) string
		parse *JSONWebTokenManager
	}{
		{"malformed", func(t *testing.T) string { return "not-a-token" }, manager},
		{"wrong secret", func(t *testing.T) string { return tokenValue }, NewJSONWebTokenManager("other", "issuer", time.Hour)},
		{"wrong issuer", func(t *testing.T) string { return tokenValue }, NewJSONWebTokenManager("secret", "other", time.Hour)},
		{"expired", func(t *testing.T) string {
			value, generationError := NewJSONWebTokenManager("secret", "issuer", -time.Hour).Generate("user-1", nil, time.Now().Add(-time.Hour))
			if generationError != nil {
				t.Fatal(generationError)
			}
			return value
		}, manager},
		{"wrong algorithm", func(t *testing.T) string {
			value, signingError := jwt.NewWithClaims(jwt.SigningMethodHS384, AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user-1", Issuer: "issuer", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString([]byte("secret"))
			if signingError != nil {
				t.Fatal(signingError)
			}
			return value
		}, manager},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims, parseError := test.parse.Parse(test.value(t))
			if claims != nil || !errors.Is(parseError, ErrInvalidAccessToken) {
				t.Fatalf("Parse() = %#v, %v", claims, parseError)
			}
		})
	}
}
