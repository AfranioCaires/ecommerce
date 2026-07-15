package security

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptPasswordHasher(t *testing.T) {
	hasher := NewBcryptPasswordHasher(bcrypt.MinCost)
	hash, err := hasher.Hash("password")
	if err != nil || hash == "" {
		t.Fatalf("Hash() = %q, %v", hash, err)
	}
	if err := hasher.Compare(hash, "password"); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if err := hasher.Compare(hash, "wrong"); err == nil {
		t.Fatal("Compare() error = nil")
	}

	invalidHasher := NewBcryptPasswordHasher(bcrypt.MaxCost + 1)
	hash, err = invalidHasher.Hash("password")
	if err == nil || hash != "" {
		t.Fatalf("Hash() = %q, %v", hash, err)
	}
}
