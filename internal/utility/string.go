package utility

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
)

// TrimStringEnd removes the last l characters from str.
func TrimStringEnd(str string, l int) string {
	if l <= 0 || len(str) == 0 {
		return str
	}
	if l >= len(str) {
		return ""
	}
	return str[:len(str)-l]
}

// IsInt reports whether s represents a valid integer.
func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// FormatJSON returns a formatted JSON string with indentation.
func FormatJSON(b any) (string, error) {
	formatted, err := json.MarshalIndent(b, "", "    ")
	if err != nil {
		return "", err
	}
	return string(formatted), nil
}

// MakeRandomString generates a cryptographically secure random string
// Returns 32 random bytes encoded as base64 (43 characters)
func MakeRandomString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of a raw token string.
// Used to store and compare tokens without keeping the plaintext in the database.
func HashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
