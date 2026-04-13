package remote

import (
	"errors"
	"testing"
)

func TestErrRemoteMode(t *testing.T) {
	if ErrRemoteMode == nil {
		t.Fatal("ErrRemoteMode should not be nil")
	}
	want := "operation not available in remote mode"
	if got := ErrRemoteMode.Error(); got != want {
		t.Errorf("ErrRemoteMode.Error() = %q, want %q", got, want)
	}
}

func TestErrNotSupported_Error(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{method: "CreateAllTables", want: "remote: method not supported: CreateAllTables"},
		{method: "DropAllTables", want: "remote: method not supported: DropAllTables"},
		{method: "", want: "remote: method not supported: "},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			err := ErrNotSupported{Method: tt.method}
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrNotSupported_ImplementsError(t *testing.T) {
	var err error = ErrNotSupported{Method: "Ping"}
	if err == nil {
		t.Fatal("ErrNotSupported should satisfy error interface")
	}
}

func TestErrNotSupported_NotWrappedByErrRemoteMode(t *testing.T) {
	notSupported := ErrNotSupported{Method: "CreateAllTables"}
	if errors.Is(notSupported, ErrRemoteMode) {
		t.Error("ErrNotSupported should not match ErrRemoteMode")
	}
}
