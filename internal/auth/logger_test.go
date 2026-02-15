package auth_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/utility"
)

// Compile-time check: *utility.Logger must satisfy auth.Logger.
// If utility.Logger's method signatures drift from the interface, this line
// fails to compile -- catching the problem before any test runs.
var _ auth.Logger = (*utility.Logger)(nil)

// ---------------------------------------------------------------------------
// spyLogger: a minimal test double that records every call through the
// auth.Logger interface. Each method appends a logEntry so tests can inspect
// exactly what was logged.
// ---------------------------------------------------------------------------

type logEntry struct {
	level   string
	message string
	err     error
	args    []any
}

type spyLogger struct {
	entries []logEntry
}

func (s *spyLogger) Debug(message string, args ...any) {
	s.entries = append(s.entries, logEntry{level: "debug", message: message, args: args})
}

func (s *spyLogger) Info(message string, args ...any) {
	s.entries = append(s.entries, logEntry{level: "info", message: message, args: args})
}

func (s *spyLogger) Warn(message string, err error, args ...any) {
	s.entries = append(s.entries, logEntry{level: "warn", message: message, err: err, args: args})
}

func (s *spyLogger) Error(message string, err error, args ...any) {
	s.entries = append(s.entries, logEntry{level: "error", message: message, err: err, args: args})
}

// Compile-time check: spyLogger satisfies auth.Logger.
var _ auth.Logger = (*spyLogger)(nil)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestLogger_DebugRecordsMessage verifies that Debug calls are recorded with
// the correct message and variadic args through the interface.
func TestLogger_DebugRecordsMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		args    []any
	}{
		{
			name:    "simple message no args",
			message: "starting operation",
			args:    nil,
		},
		{
			name:    "message with string arg",
			message: "user %s connected",
			args:    []any{"alice"},
		},
		{
			name:    "message with multiple args",
			message: "processed %d items in %s",
			args:    []any{42, "3.2s"},
		},
		{
			name:    "empty message",
			message: "",
			args:    nil,
		},
		{
			name:    "message with nil arg",
			message: "value is %v",
			args:    []any{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyLogger{}
			var logger auth.Logger = spy

			logger.Debug(tt.message, tt.args...)

			if len(spy.entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(spy.entries))
			}
			entry := spy.entries[0]
			if entry.level != "debug" {
				t.Errorf("level = %q, want %q", entry.level, "debug")
			}
			if entry.message != tt.message {
				t.Errorf("message = %q, want %q", entry.message, tt.message)
			}
			if len(entry.args) != len(tt.args) {
				t.Fatalf("args length = %d, want %d", len(entry.args), len(tt.args))
			}
			for i, arg := range entry.args {
				if fmt.Sprintf("%v", arg) != fmt.Sprintf("%v", tt.args[i]) {
					t.Errorf("args[%d] = %v, want %v", i, arg, tt.args[i])
				}
			}
		})
	}
}

// TestLogger_InfoRecordsMessage verifies that Info calls are recorded correctly.
func TestLogger_InfoRecordsMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		args    []any
	}{
		{
			name:    "simple info message",
			message: "server started on port 8080",
			args:    nil,
		},
		{
			name:    "info with args",
			message: "serving %d routes for %s",
			args:    []any{15, "api.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyLogger{}
			var logger auth.Logger = spy

			logger.Info(tt.message, tt.args...)

			if len(spy.entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(spy.entries))
			}
			entry := spy.entries[0]
			if entry.level != "info" {
				t.Errorf("level = %q, want %q", entry.level, "info")
			}
			if entry.message != tt.message {
				t.Errorf("message = %q, want %q", entry.message, tt.message)
			}
		})
	}
}

// TestLogger_WarnRecordsMessageAndError verifies that Warn passes both the
// message and the error parameter through the interface correctly.
func TestLogger_WarnRecordsMessageAndError(t *testing.T) {
	t.Parallel()

	someErr := errors.New("disk nearly full")

	tests := []struct {
		name    string
		message string
		err     error
		args    []any
	}{
		{
			name:    "warn with error",
			message: "storage running low",
			err:     someErr,
			args:    nil,
		},
		{
			// user_provision.go line 177 passes nil as the error to Warn.
			// This verifies the interface handles nil errors gracefully.
			name:    "warn with nil error",
			message: "OAuth provider did not provide sub",
			err:     nil,
			args:    nil,
		},
		{
			name:    "warn with error and args",
			message: "retrying request to %s",
			err:     fmt.Errorf("connection refused"),
			args:    []any{"api.github.com"},
		},
		{
			name:    "warn with wrapped error",
			message: "operation failed",
			err:     fmt.Errorf("parse config: %w", errors.New("unexpected EOF")),
			args:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyLogger{}
			var logger auth.Logger = spy

			logger.Warn(tt.message, tt.err, tt.args...)

			if len(spy.entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(spy.entries))
			}
			entry := spy.entries[0]
			if entry.level != "warn" {
				t.Errorf("level = %q, want %q", entry.level, "warn")
			}
			if entry.message != tt.message {
				t.Errorf("message = %q, want %q", entry.message, tt.message)
			}

			// Check error equality
			if tt.err == nil && entry.err != nil {
				t.Errorf("err = %v, want nil", entry.err)
			}
			if tt.err != nil && entry.err == nil {
				t.Errorf("err = nil, want %v", tt.err)
			}
			if tt.err != nil && entry.err != nil && entry.err.Error() != tt.err.Error() {
				t.Errorf("err = %q, want %q", entry.err.Error(), tt.err.Error())
			}
		})
	}
}

// TestLogger_ErrorRecordsMessageAndError verifies that Error passes both the
// message and the error parameter through the interface correctly.
func TestLogger_ErrorRecordsMessageAndError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message string
		err     error
		args    []any
	}{
		{
			name:    "error with error value",
			message: "token refresh API call failed",
			err:     errors.New("401 unauthorized"),
			args:    nil,
		},
		{
			name:    "error with nil error",
			message: "unexpected state",
			err:     nil,
			args:    nil,
		},
		{
			name:    "error with args and error",
			message: "failed to create user %s",
			err:     fmt.Errorf("duplicate key"),
			args:    []any{"alice@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyLogger{}
			var logger auth.Logger = spy

			logger.Error(tt.message, tt.err, tt.args...)

			if len(spy.entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(spy.entries))
			}
			entry := spy.entries[0]
			if entry.level != "error" {
				t.Errorf("level = %q, want %q", entry.level, "error")
			}
			if entry.message != tt.message {
				t.Errorf("message = %q, want %q", entry.message, tt.message)
			}

			if tt.err == nil && entry.err != nil {
				t.Errorf("err = %v, want nil", entry.err)
			}
			if tt.err != nil && entry.err == nil {
				t.Errorf("err = nil, want %v", tt.err)
			}
			if tt.err != nil && entry.err != nil && entry.err.Error() != tt.err.Error() {
				t.Errorf("err = %q, want %q", entry.err.Error(), tt.err.Error())
			}
		})
	}
}

// TestLogger_SequentialCallsAccumulate verifies that multiple calls through
// the interface accumulate in order. This matters because TokenRefresher and
// UserProvisioner make multiple log calls during a single operation.
func TestLogger_SequentialCallsAccumulate(t *testing.T) {
	t.Parallel()

	spy := &spyLogger{}
	var logger auth.Logger = spy

	someErr := errors.New("timeout")

	logger.Debug("step 1")
	logger.Info("step 2", "extra")
	logger.Warn("step 3", someErr)
	logger.Error("step 4", someErr, "detail")

	if len(spy.entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(spy.entries))
	}

	expectedLevels := []string{"debug", "info", "warn", "error"}
	expectedMessages := []string{"step 1", "step 2", "step 3", "step 4"}

	for i, entry := range spy.entries {
		if entry.level != expectedLevels[i] {
			t.Errorf("entries[%d].level = %q, want %q", i, entry.level, expectedLevels[i])
		}
		if entry.message != expectedMessages[i] {
			t.Errorf("entries[%d].message = %q, want %q", i, entry.message, expectedMessages[i])
		}
	}

	// Warn and Error should carry the error
	if spy.entries[2].err == nil {
		t.Error("entries[2] (warn) should have an error, got nil")
	}
	if spy.entries[3].err == nil {
		t.Error("entries[3] (error) should have an error, got nil")
	}

	// Debug and Info should NOT carry an error
	if spy.entries[0].err != nil {
		t.Errorf("entries[0] (debug) should have nil error, got %v", spy.entries[0].err)
	}
	if spy.entries[1].err != nil {
		t.Errorf("entries[1] (info) should have nil error, got %v", spy.entries[1].err)
	}
}

// TestLogger_InterfaceAssignment verifies that the Logger interface can be
// assigned to and called through a variable, which is how TokenRefresher and
// UserProvisioner use it (stored as a struct field of type Logger).
func TestLogger_InterfaceAssignment(t *testing.T) {
	t.Parallel()

	spy := &spyLogger{}

	// Simulate what NewTokenRefresher / NewUserProvisioner do:
	// store a Logger interface value in a struct field
	type holder struct {
		log auth.Logger
	}

	h := holder{log: spy}

	h.log.Debug("test from holder")
	h.log.Info("info from holder")
	h.log.Warn("warn from holder", errors.New("oops"))
	h.log.Error("error from holder", errors.New("bad"))

	if len(spy.entries) != 4 {
		t.Fatalf("expected 4 entries from holder usage, got %d", len(spy.entries))
	}
}

// TestLogger_VariadicArgsPreserveTypes verifies that variadic arguments of
// different types pass through the interface without type loss. Callers in
// token_refresh.go and user_provision.go pass types.UserID (a string type),
// time.Duration, and plain strings as variadic args.
func TestLogger_VariadicArgsPreserveTypes(t *testing.T) {
	t.Parallel()

	spy := &spyLogger{}
	var logger auth.Logger = spy

	type customID string
	id := customID("01HXYZ123")

	logger.Debug("user %s did %d things in %v", id, 42, "3s")

	if len(spy.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(spy.entries))
	}

	args := spy.entries[0].args
	if len(args) != 3 {
		t.Fatalf("args length = %d, want 3", len(args))
	}

	// First arg should be the customID value
	if gotID, ok := args[0].(customID); !ok {
		t.Errorf("args[0] type = %T, want customID", args[0])
	} else if gotID != id {
		t.Errorf("args[0] = %v, want %v", gotID, id)
	}

	// Second arg should be int
	if gotInt, ok := args[1].(int); !ok {
		t.Errorf("args[1] type = %T, want int", args[1])
	} else if gotInt != 42 {
		t.Errorf("args[1] = %d, want 42", gotInt)
	}

	// Third arg should be string
	if gotStr, ok := args[2].(string); !ok {
		t.Errorf("args[2] type = %T, want string", args[2])
	} else if gotStr != "3s" {
		t.Errorf("args[2] = %q, want %q", gotStr, "3s")
	}
}

// TestLogger_ZeroValueArgs verifies behavior with zero-value and edge-case
// arguments. The Logger interface uses ...any which can receive anything
// including nil, empty strings, and zero integers.
func TestLogger_ZeroValueArgs(t *testing.T) {
	t.Parallel()

	spy := &spyLogger{}
	var logger auth.Logger = spy

	logger.Debug("zero values", "", 0, false, nil)

	if len(spy.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(spy.entries))
	}

	args := spy.entries[0].args
	if len(args) != 4 {
		t.Fatalf("args length = %d, want 4", len(args))
	}

	if args[0] != "" {
		t.Errorf("args[0] = %v, want empty string", args[0])
	}
	if args[1] != 0 {
		t.Errorf("args[1] = %v, want 0", args[1])
	}
	if args[2] != false {
		t.Errorf("args[2] = %v, want false", args[2])
	}
	if args[3] != nil {
		t.Errorf("args[3] = %v, want nil", args[3])
	}
}

// TestLogger_NoArgsProducesEmptySlice verifies that calling with no variadic
// args produces an empty (or nil) args slice, not a slice with one nil element.
func TestLogger_NoArgsProducesEmptySlice(t *testing.T) {
	t.Parallel()

	spy := &spyLogger{}
	var logger auth.Logger = spy

	logger.Debug("no args")
	logger.Info("no args either")
	logger.Warn("still no args", errors.New("err"))
	logger.Error("same", errors.New("err"))

	for i, entry := range spy.entries {
		if len(entry.args) != 0 {
			t.Errorf("entries[%d].args length = %d, want 0", i, len(entry.args))
		}
	}
}
