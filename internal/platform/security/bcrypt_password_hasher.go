package security

import "golang.org/x/crypto/bcrypt"

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: cost}
}

func (passwordHasher *BcryptPasswordHasher) Hash(
	password string,
) (string, error) {
	passwordHash, errorValue := bcrypt.GenerateFromPassword(
		[]byte(password),
		passwordHasher.cost,
	)
	if errorValue != nil {
		return "", errorValue
	}

	return string(passwordHash), nil
}

func (passwordHasher *BcryptPasswordHasher) Compare(
	passwordHash string,
	password string,
) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
}
