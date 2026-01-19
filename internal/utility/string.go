package utility

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strconv"
)

func TrimStringEnd(str string, l int) string {
	if len(str) > 0 {
		newStr := str[:len(str)-l]
		return newStr
	} else {
		return str
	}
}

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

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

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
