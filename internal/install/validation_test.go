package install_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/install"
)

func TestValidatePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{name: "valid port 8080", input: "8080", wantErr: false},
		{name: "valid port lower bound", input: "1024", wantErr: false},
		{name: "valid port upper bound", input: "65535", wantErr: false},
		{name: "valid port mid range", input: "3000", wantErr: false},
		{name: "empty string", input: "", wantErr: true, errMsg: "port cannot be empty"},
		{name: "non-numeric", input: "abc", wantErr: true, errMsg: "port must be a number"},
		{name: "below range", input: "1023", wantErr: true, errMsg: "port must be between 1024 and 65535"},
		{name: "above range", input: "65536", wantErr: true, errMsg: "port must be between 1024 and 65535"},
		{name: "zero", input: "0", wantErr: true, errMsg: "port must be between 1024 and 65535"},
		{name: "negative", input: "-1", wantErr: true, errMsg: "port must be between 1024 and 65535"},
		{name: "port with spaces", input: " 8080 ", wantErr: true, errMsg: "port must be a number"},
		{name: "float", input: "80.80", wantErr: true, errMsg: "port must be a number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidatePort(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePortOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty string is valid", input: "", wantErr: false},
		{name: "valid port", input: "8080", wantErr: false},
		{name: "invalid port", input: "abc", wantErr: true},
		{name: "below range", input: "80", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidatePortOrEmpty(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePortOrEmpty(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{name: "simple hostname", input: "localhost", wantErr: false},
		{name: "hostname with port", input: "localhost:8080", wantErr: false},
		{name: "full http URL", input: "http://example.com", wantErr: false},
		{name: "full https URL", input: "https://example.com", wantErr: false},
		{name: "URL with path", input: "https://example.com/api", wantErr: false},
		{name: "empty string", input: "", wantErr: true, errMsg: "URL cannot be empty"},
		{name: "scheme only no host", input: "http://", wantErr: true, errMsg: "URL must have a host"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidateURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDBPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{name: "valid .db path", input: "test.db", wantErr: false},
		{name: "valid nested .db path", input: "/data/modula.db", wantErr: false},
		{name: "empty string", input: "", wantErr: true, errMsg: "database path cannot be empty"},
		{name: "wrong extension .sqlite", input: "test.sqlite", wantErr: true, errMsg: "should have .db extension"},
		{name: "no extension", input: "testdb", wantErr: true, errMsg: "should have .db extension"},
		{name: "wrong extension .sql", input: "test.sql", wantErr: true, errMsg: "should have .db extension"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidateDBPath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fieldName string
		input     string
		wantErr   bool
	}{
		{name: "non-empty value", fieldName: "username", input: "admin", wantErr: false},
		{name: "empty string", fieldName: "username", input: "", wantErr: true},
		{name: "whitespace only", fieldName: "email", input: "   ", wantErr: true},
		{name: "tabs only", fieldName: "email", input: "\t\t", wantErr: true},
		{name: "value with spaces", fieldName: "name", input: "John Doe", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validator := install.ValidateNotEmpty(tt.fieldName)
			err := validator(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.fieldName) {
					t.Errorf("error = %q, should mention field name %q", err.Error(), tt.fieldName)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDBName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple name", input: "modula", wantErr: false},
		{name: "with underscore", input: "modula_cms", wantErr: false},
		{name: "with numbers", input: "db123", wantErr: false},
		{name: "uppercase", input: "ModulaCMS", wantErr: false},
		{name: "all underscore", input: "___", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "with hyphen", input: "modula-cms", wantErr: true},
		{name: "with space", input: "modula cms", wantErr: true},
		{name: "with dot", input: "modula.cms", wantErr: true},
		{name: "with semicolon", input: "db;drop", wantErr: true},
		{name: "with slash", input: "db/name", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidateDBName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDBName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCookieName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple name", input: "session", wantErr: false},
		{name: "with underscore", input: "session_id", wantErr: false},
		{name: "with hyphen", input: "session-id", wantErr: false},
		{name: "with numbers", input: "sess123", wantErr: false},
		{name: "uppercase", input: "SESSION_ID", wantErr: false},
		{name: "empty", input: "", wantErr: true},
		{name: "with space", input: "session id", wantErr: true},
		{name: "with dot", input: "session.id", wantErr: true},
		{name: "with semicolon", input: "session;id", wantErr: true},
		{name: "with equals", input: "session=val", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidateCookieName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCookieName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{name: "valid 8 chars", input: "abcd1234", wantErr: false},
		{name: "valid long password", input: "a-very-secure-password-2026", wantErr: false},
		{name: "exactly 72 bytes", input: strings.Repeat("a", 72), wantErr: false},
		{name: "too short 7 chars", input: "abcd123", wantErr: true, errMsg: "at least 8 characters"},
		{name: "empty", input: "", wantErr: true, errMsg: "at least 8 characters"},
		{name: "73 bytes exceeds bcrypt limit", input: strings.Repeat("a", 73), wantErr: true, errMsg: "must not exceed 72 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := install.ValidatePassword(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	t.Parallel()

	t.Run("valid path in existing directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")
		err := install.ValidateConfigPath(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid path when file already exists", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "modula.config.json")
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		err := install.ValidateConfigPath(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		err := install.ValidateConfigPath("")
		if err == nil {
			t.Fatal("expected error for empty path, got nil")
		}
	})

	t.Run("directory does not exist", func(t *testing.T) {
		t.Parallel()
		err := install.ValidateConfigPath("/nonexistent/dir/config.json")
		if err == nil {
			t.Fatal("expected error for nonexistent directory, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error = %q, want substring %q", err.Error(), "does not exist")
		}
	})

	t.Run("parent path is a file not directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		fakedir := filepath.Join(dir, "notadir")
		if err := os.WriteFile(fakedir, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		err := install.ValidateConfigPath(filepath.Join(fakedir, "config.json"))
		if err == nil {
			t.Fatal("expected error when parent is a file, got nil")
		}
		if !strings.Contains(err.Error(), "is not a directory") {
			t.Errorf("error = %q, want substring %q", err.Error(), "is not a directory")
		}
	})
}

func TestValidateDirPath(t *testing.T) {
	t.Parallel()

	t.Run("valid existing directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		err := install.ValidateDirPath(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		err := install.ValidateDirPath("")
		if err == nil {
			t.Fatal("expected error for empty path, got nil")
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()
		err := install.ValidateDirPath("/nonexistent/path")
		if err == nil {
			t.Fatal("expected error for nonexistent path, got nil")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error = %q, want substring %q", err.Error(), "does not exist")
		}
	})

	t.Run("path is a file not directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f := filepath.Join(dir, "afile")
		if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		err := install.ValidateDirPath(f)
		if err == nil {
			t.Fatal("expected error when path is a file, got nil")
		}
		if !strings.Contains(err.Error(), "is not a directory") {
			t.Errorf("error = %q, want substring %q", err.Error(), "is not a directory")
		}
	})
}
