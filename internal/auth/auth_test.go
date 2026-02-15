// White-box tests for auth.go: password hashing and verification.
package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// ---------------------------------------------------------------------------
// HashPassword
// ---------------------------------------------------------------------------

func TestHashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "typical password",
			password: "correcthorsebatterystaple",
		},
		{
			name:     "short password",
			password: "a",
		},
		{
			name:     "password with special characters",
			password: "p@$$w0rd!#%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:     "unicode password",
			password: "Mot de passe", // French with spaces
		},
		{
			name:     "password with emoji",
			password: "hunter2\U0001F512secure",
		},
		{
			name:     "password with newlines and tabs",
			password: "line1\nline2\ttab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword(%q) returned error: %v", tt.password, err)
			}
			if hash == "" {
				t.Fatal("HashPassword returned empty hash")
			}
			// Hash must not equal the plaintext
			if hash == tt.password {
				t.Error("hash equals plaintext password")
			}
			// Must be valid bcrypt
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password)); err != nil {
				t.Errorf("bcrypt.CompareHashAndPassword failed for valid hash: %v", err)
			}
		})
	}
}

func TestHashPassword_RejectsEmptyPassword(t *testing.T) {
	t.Parallel()

	_, err := HashPassword("")
	if err == nil {
		t.Fatal("expected error for empty password, got nil")
	}
	if !strings.Contains(err.Error(), "must not be empty") {
		t.Errorf("error = %q, want it to contain 'must not be empty'", err.Error())
	}
}

func TestHashPassword_CostIs12(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("testcost")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost failed: %v", err)
	}
	if cost != 12 {
		t.Errorf("bcrypt cost = %d, want 12", cost)
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	t.Parallel()

	// bcrypt uses a random salt, so the same password should produce different hashes
	hash1, err := HashPassword("duplicate")
	if err != nil {
		t.Fatalf("first HashPassword call failed: %v", err)
	}
	hash2, err := HashPassword("duplicate")
	if err != nil {
		t.Fatalf("second HashPassword call failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("two calls with same password produced identical hashes; salt is not random")
	}
}

// Go's bcrypt implementation rejects passwords exceeding 72 bytes with an error.
// Our explicit check provides a clearer error message.
func TestHashPassword_RejectsOver72Bytes(t *testing.T) {
	t.Parallel()

	password := strings.Repeat("A", 73)
	_, err := HashPassword(password)
	if err == nil {
		t.Fatal("expected error for password exceeding 72 bytes, got nil")
	}
	if !strings.Contains(err.Error(), "must not exceed 72 bytes") {
		t.Errorf("error = %q, want it to contain 'must not exceed 72 bytes'", err.Error())
	}
}

// Exactly 72 bytes should succeed -- the boundary case.
func TestHashPassword_Accepts72Bytes(t *testing.T) {
	t.Parallel()

	password := strings.Repeat("A", 72)
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(%d bytes) returned error: %v", len(password), err)
	}
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash failed for 72-byte password")
	}
}

// ---------------------------------------------------------------------------
// CheckPasswordHash
// ---------------------------------------------------------------------------

func TestCheckPasswordHash(t *testing.T) {
	t.Parallel()

	// Pre-generate a hash for the known password
	knownPassword := "securepassword123"
	knownHash, err := HashPassword(knownPassword)
	if err != nil {
		t.Fatalf("setup: HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: knownPassword,
			hash:     knownHash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     knownHash,
			want:     false,
		},
		{
			name:     "empty password against valid hash",
			password: "",
			hash:     knownHash,
			want:     false,
		},
		{
			name:     "valid password against empty hash",
			password: knownPassword,
			hash:     "",
			want:     false,
		},
		{
			name:     "empty password against empty hash",
			password: "",
			hash:     "",
			want:     false,
		},
		{
			name:     "valid password against corrupted hash",
			password: knownPassword,
			hash:     "not-a-bcrypt-hash",
			want:     false,
		},
		{
			name:     "valid password against truncated hash",
			password: knownPassword,
			hash:     knownHash[:10],
			want:     false,
		},
		{
			name:     "case-sensitive password mismatch",
			password: "Securepassword123", // capital S
			hash:     knownHash,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CheckPasswordHash(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPasswordHash(%q, hash) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// HashPassword + CheckPasswordHash round-trip
// ---------------------------------------------------------------------------

func TestHashAndCheck_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
	}{
		{name: "alphanumeric", password: "password123"},
		{name: "spaces only", password: "   "},
		{name: "max length password (72 bytes)", password: strings.Repeat("x", 72)},
		{name: "unicode CJK", password: "\u5bc6\u7801\u6d4b\u8bd5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword(%q) failed: %v", tt.password, err)
			}

			if !CheckPasswordHash(tt.password, hash) {
				t.Errorf("CheckPasswordHash(%q, hash) = false, want true", tt.password)
			}

			// A different password should not match.
			// We prepend "WRONG_" so the difference is within the first 72 bytes
			// (bcrypt truncates at 72 bytes, so appending would be invisible for
			// max-length passwords).
			wrongPassword := "WRONG_" + tt.password
			if len(wrongPassword) > 72 {
				wrongPassword = wrongPassword[:72]
			}
			if CheckPasswordHash(wrongPassword, hash) {
				t.Errorf("CheckPasswordHash(wrongPassword, hash) = true, want false")
			}
		})
	}
}
