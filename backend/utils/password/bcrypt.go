package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost // 10 — safe and fast enough for most deployments

// HashPassword returns the bcrypt hash of the plain-text password.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plain-text password against a stored bcrypt hash.
// Returns nil if they match, an error otherwise.
func CheckPassword(plain, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
