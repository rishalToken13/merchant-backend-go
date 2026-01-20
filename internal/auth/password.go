package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const BcryptCost = bcrypt.DefaultCost

func HashPassword(plain string) (string, error) {
	if len(plain) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt generate: %w", err)
	}
	return string(hashBytes), nil
}

func VerifyPassword(hash string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}