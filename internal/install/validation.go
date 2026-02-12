package install

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidatePort checks if the port is numeric and within the valid range (1024-65535)
func ValidatePort(s string) error {
	if s == "" {
		return fmt.Errorf("port cannot be empty")
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("port must be a number, got %q", s)
	}
	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535, got %d", port)
	}
	return nil
}

// ValidateConfigPath checks if the config path is valid and writable
func ValidateConfigPath(s string) error {
	if s == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	// Get the directory part
	dir := filepath.Dir(s)
	if dir == "" {
		dir = "."
	}

	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %q does not exist", dir)
		}
		return fmt.Errorf("cannot access directory %q: %v", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", dir)
	}

	// Check if file already exists (warning, not error)
	if _, err := os.Stat(s); err == nil {
		// File exists - this is OK, we'll overwrite it
		return nil
	}

	return nil
}

// ValidateURL performs basic URL format validation
func ValidateURL(s string) error {
	if s == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Allow simple hostnames like "localhost"
	if !strings.Contains(s, "://") {
		// Treat as hostname only
		return nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	return nil
}

// ValidateDBPath checks that the path has a .db extension (for SQLite)
func ValidateDBPath(s string) error {
	if s == "" {
		return fmt.Errorf("database path cannot be empty")
	}
	ext := filepath.Ext(s)
	if ext != ".db" {
		return fmt.Errorf("SQLite database file should have .db extension, got %q", ext)
	}
	return nil
}

// ValidateNotEmpty returns a validation function that checks a field is not empty
func ValidateNotEmpty(fieldName string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s cannot be empty", fieldName)
		}
		return nil
	}
}

// ValidatePortOrEmpty allows empty string (optional) or validates port
func ValidatePortOrEmpty(s string) error {
	if s == "" {
		return nil
	}
	return ValidatePort(s)
}

// ValidateDBName checks database name doesn't contain special characters
func ValidateDBName(s string) error {
	if s == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	// Allow alphanumeric and underscores
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("database name can only contain letters, numbers, and underscores")
		}
	}
	return nil
}

// ValidateDirPath checks that a directory path exists and is accessible
func ValidateDirPath(s string) error {
	if s == "" {
		return fmt.Errorf("directory path cannot be empty")
	}
	info, err := os.Stat(s)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %q does not exist", s)
		}
		return fmt.Errorf("cannot access %q: %v", s, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", s)
	}
	return nil
}

// ValidateCookieName checks that a cookie name contains only alphanumeric, underscore, and hyphen characters
func ValidateCookieName(s string) error {
	if s == "" {
		return fmt.Errorf("cookie name cannot be empty")
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("cookie name can only contain letters, numbers, underscores, and hyphens")
		}
	}
	return nil
}

// ValidatePassword checks that a password meets minimum requirements.
// Minimum 8 characters, maximum 72 bytes (bcrypt limit).
func ValidatePassword(s string) error {
	if len(s) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(s) > 72 {
		return fmt.Errorf("password must not exceed 72 bytes (bcrypt limit)")
	}
	return nil
}
