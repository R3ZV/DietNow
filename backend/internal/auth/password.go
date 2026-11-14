package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 8
	// MaxPasswordBytes is bcrypt's input limit; longer inputs are rejected, not truncated.
	MaxPasswordBytes = 72
)

var ErrWeakPassword = errors.New("password must be 8-72 bytes long")

// HashPassword returns a salted bcrypt hash.
func HashPassword(password string, cost int) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func ValidatePassword(password string) error {
	if len([]rune(password)) < MinPasswordLen || len(password) > MaxPasswordBytes {
		return ErrWeakPassword
	}
	return nil
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
