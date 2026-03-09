package utility

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================
// TrimStringEnd
// ============================================================

func TestTrimStringEnd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		str  string
		l    int
		want string
	}{
		{name: "trim last char", str: "hello", l: 1, want: "hell"},
		{name: "trim last 3 chars", str: "hello", l: 3, want: "he"},
		{name: "trim all chars", str: "hello", l: 5, want: ""},
		{name: "trim nothing", str: "hello", l: 0, want: "hello"},
		{name: "empty string returns empty", str: "", l: 0, want: ""},
		{name: "empty string with nonzero l returns empty", str: "", l: 3, want: ""},
		{name: "single char trim one", str: "x", l: 1, want: ""},
		{name: "l exceeds length", str: "hi", l: 5, want: ""},
		{name: "l equals length", str: "hi", l: 2, want: ""},
		{name: "negative l returns unchanged", str: "hello", l: -3, want: "hello"},
		// "cafe\u0301" is 6 bytes: c(1) a(1) f(1) e(1) \u0301(2). Trimming 2 bytes removes the combining accent.
		{name: "unicode string trims bytes not runes", str: "cafe\u0301", l: 2, want: "cafe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TrimStringEnd(tt.str, tt.l)
			if got != tt.want {
				t.Errorf("TrimStringEnd(%q, %d) = %q, want %q", tt.str, tt.l, got, tt.want)
			}
		})
	}
}

// ============================================================
// IsInt
// ============================================================

func TestIsInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{name: "positive integer", s: "42", want: true},
		{name: "zero", s: "0", want: true},
		{name: "negative integer", s: "-7", want: true},
		{name: "large number", s: "2147483647", want: true},
		{name: "empty string", s: "", want: false},
		{name: "float", s: "3.14", want: false},
		{name: "letters", s: "abc", want: false},
		{name: "mixed", s: "12abc", want: false},
		{name: "whitespace", s: " 5 ", want: false},
		{name: "hex notation", s: "0xFF", want: false},
		{name: "leading plus", s: "+5", want: true}, // strconv.Atoi accepts leading +
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsInt(tt.s)
			if got != tt.want {
				t.Errorf("IsInt(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// ============================================================
// FormatJSON
// ============================================================

func TestFormatJSON(t *testing.T) {
	t.Parallel()

	t.Run("formats struct with indentation", func(t *testing.T) {
		t.Parallel()
		input := map[string]string{"key": "value"}
		got, err := FormatJSON(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify it is valid JSON
		var parsed map[string]string
		if err := json.Unmarshal([]byte(got), &parsed); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %s", err, got)
		}
		if parsed["key"] != "value" {
			t.Errorf("parsed key = %q, want %q", parsed["key"], "value")
		}
		// Verify indentation (4 spaces)
		if !strings.Contains(got, "    ") {
			t.Errorf("expected 4-space indentation in output:\n%s", got)
		}
	})

	t.Run("formats nil", func(t *testing.T) {
		t.Parallel()
		got, err := FormatJSON(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "null" {
			t.Errorf("FormatJSON(nil) = %q, want %q", got, "null")
		}
	})

	t.Run("formats nested structure", func(t *testing.T) {
		t.Parallel()
		input := map[string]any{
			"name": "test",
			"nested": map[string]int{
				"a": 1,
			},
		}
		got, err := FormatJSON(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "\"name\"") {
			t.Errorf("expected key 'name' in output:\n%s", got)
		}
	})

	t.Run("returns error for unmarshalable input", func(t *testing.T) {
		t.Parallel()
		// Functions cannot be marshaled to JSON
		_, err := FormatJSON(func() {})
		if err == nil {
			t.Fatal("expected error for unmarshalable input, got nil")
		}
	})
}

// ============================================================
// MakeRandomString
// ============================================================

func TestMakeRandomString(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty string", func(t *testing.T) {
		t.Parallel()
		got, err := MakeRandomString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == "" {
			t.Fatal("MakeRandomString() returned empty string")
		}
	})

	t.Run("returns base64url encoded 43 chars", func(t *testing.T) {
		t.Parallel()
		got, err := MakeRandomString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 32 bytes base64url-encoded without padding = 43 characters
		if len(got) != 43 {
			t.Errorf("MakeRandomString() length = %d, want 43", len(got))
		}
	})

	t.Run("produces unique values", func(t *testing.T) {
		t.Parallel()
		a, err := MakeRandomString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, err := MakeRandomString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a == b {
			t.Error("MakeRandomString() produced identical values on consecutive calls")
		}
	})
}

// ============================================================
// IsValidEmail
// ============================================================

func TestIsValidEmail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "valid simple email", email: "user@example.com", want: true},
		{name: "valid with dots", email: "first.last@example.com", want: true},
		{name: "valid with plus", email: "user+tag@example.com", want: true},
		{name: "valid with subdomain", email: "user@mail.example.com", want: true},
		{name: "valid with numbers", email: "user123@example456.com", want: true},
		{name: "valid with hyphen domain", email: "user@my-domain.com", want: true},
		{name: "valid with percent", email: "user%name@example.com", want: true},
		{name: "valid two-char TLD", email: "user@example.uk", want: true},
		{name: "empty string", email: "", want: false},
		{name: "no at sign", email: "userexample.com", want: false},
		{name: "no domain", email: "user@", want: false},
		{name: "no local part", email: "@example.com", want: false},
		{name: "no TLD", email: "user@example", want: false},
		{name: "double at", email: "user@@example.com", want: false},
		{name: "space in address", email: "user @example.com", want: false},
		{name: "single char TLD", email: "user@example.c", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}
