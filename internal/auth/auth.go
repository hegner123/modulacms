package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password must not be empty")
	}
	if len(password) > 72 {
		return "", fmt.Errorf("password must not exceed 72 bytes, got %d", len(password))
	}
	// Use cost of 12 (which is a good balance of security and performance)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
