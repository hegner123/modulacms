package install_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/install"
)

func TestInstallError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantSubs []string
	}{
		{
			name:     "with hint includes hint line",
			err:      install.ErrConfigWrite(fmt.Errorf("permission denied"), "/etc/config.json"),
			wantSubs: []string{"write config file", "permission denied", "Hint:", "/etc/config.json"},
		},
		{
			name:     "without hint omits hint line",
			err:      install.ErrValidation("port", fmt.Errorf("must be numeric")),
			wantSubs: []string{"validate port", "must be numeric"},
		},
		{
			name:     "user aborted has no hint",
			err:      install.ErrUserAborted(),
			wantSubs: []string{"installation", "cancelled by user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			msg := tt.err.Error()
			for _, sub := range tt.wantSubs {
				if !strings.Contains(msg, sub) {
					t.Errorf("Error() = %q, missing substring %q", msg, sub)
				}
			}
		})
	}
}

func TestInstallError_ErrorWithoutHint(t *testing.T) {
	t.Parallel()

	err := install.ErrValidation("email", fmt.Errorf("invalid format"))
	msg := err.Error()
	if strings.Contains(msg, "Hint:") {
		t.Errorf("Error() = %q, should not contain 'Hint:' when hint is empty", msg)
	}
}

func TestInstallError_Unwrap(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("underlying failure")
	installErr := install.ErrDBTables(cause)

	unwrapped := errors.Unwrap(installErr)
	if unwrapped == nil {
		t.Fatal("Unwrap() returned nil, want underlying error")
	}
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestInstallError_ErrorsIs(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("db locked")
	installErr := install.ErrDBBootstrap(cause)

	if !errors.Is(installErr, cause) {
		t.Error("errors.Is should find the wrapped cause")
	}
}

func TestErrConfigWrite(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("disk full")
	err := install.ErrConfigWrite(cause, "/tmp/config.json")
	msg := err.Error()

	if !strings.Contains(msg, "write config file") {
		t.Errorf("missing operation in error: %q", msg)
	}
	if !strings.Contains(msg, "disk full") {
		t.Errorf("missing cause in error: %q", msg)
	}
	if !strings.Contains(msg, "/tmp/config.json") {
		t.Errorf("missing path in hint: %q", msg)
	}
}

func TestErrDBConnect_SQLite(t *testing.T) {
	t.Parallel()

	err := install.ErrDBConnect(fmt.Errorf("open failed"), "sqlite")
	msg := err.Error()
	if !strings.Contains(msg, ".db extension") {
		t.Errorf("sqlite hint should mention .db extension, got: %q", msg)
	}
}

func TestErrDBConnect_NonSQLite(t *testing.T) {
	t.Parallel()

	err := install.ErrDBConnect(fmt.Errorf("conn refused"), "postgres")
	msg := err.Error()
	if !strings.Contains(msg, "database server is running") {
		t.Errorf("non-sqlite hint should mention server running, got: %q", msg)
	}
}

func TestErrDBTables(t *testing.T) {
	t.Parallel()

	err := install.ErrDBTables(fmt.Errorf("access denied"))
	msg := err.Error()
	if !strings.Contains(msg, "create database tables") {
		t.Errorf("missing operation: %q", msg)
	}
	if !strings.Contains(msg, "CREATE TABLE") {
		t.Errorf("hint should mention CREATE TABLE: %q", msg)
	}
}

func TestErrDBBootstrap(t *testing.T) {
	t.Parallel()

	err := install.ErrDBBootstrap(fmt.Errorf("duplicate key"))
	msg := err.Error()
	if !strings.Contains(msg, "insert bootstrap data") {
		t.Errorf("missing operation: %q", msg)
	}
	if !strings.Contains(msg, "fresh database") {
		t.Errorf("hint should mention fresh database: %q", msg)
	}
}

func TestErrBucketConnect(t *testing.T) {
	t.Parallel()

	err := install.ErrBucketConnect(fmt.Errorf("timeout"))
	msg := err.Error()
	if !strings.Contains(msg, "S3 bucket") {
		t.Errorf("missing operation: %q", msg)
	}
	if !strings.Contains(msg, "access key") {
		t.Errorf("hint should mention access key: %q", msg)
	}
}

func TestErrMaxRetries(t *testing.T) {
	t.Parallel()

	err := install.ErrMaxRetries(3)
	msg := err.Error()
	if !strings.Contains(msg, "3") {
		t.Errorf("should contain attempt count: %q", msg)
	}
	if !strings.Contains(msg, "--install") {
		t.Errorf("hint should mention --install flag: %q", msg)
	}
}

func TestErrUserAborted(t *testing.T) {
	t.Parallel()

	err := install.ErrUserAborted()
	msg := err.Error()
	if !strings.Contains(msg, "cancelled by user") {
		t.Errorf("should contain cancellation message: %q", msg)
	}
}
