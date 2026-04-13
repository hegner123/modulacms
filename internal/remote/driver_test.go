package remote

import (
	"context"
	"errors"
	"fmt"
	"testing"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// RemoteStatus
// ---------------------------------------------------------------------------

func TestRemoteStatus_String(t *testing.T) {
	tests := []struct {
		status RemoteStatus
		want   string
	}{
		{StatusUnknown, "unknown"},
		{StatusConnected, "connected"},
		{StatusDisconnected, "disconnected"},
		{RemoteStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRemoteStatus_Constants(t *testing.T) {
	if StatusUnknown != 0 {
		t.Errorf("StatusUnknown = %d, want 0", StatusUnknown)
	}
	if StatusConnected != 1 {
		t.Errorf("StatusConnected = %d, want 1", StatusConnected)
	}
	if StatusDisconnected != 2 {
		t.Errorf("StatusDisconnected = %d, want 2", StatusDisconnected)
	}
}

// ---------------------------------------------------------------------------
// newTestDriver creates a RemoteDriver without the health check (no network)
// ---------------------------------------------------------------------------

func newTestDriver() *RemoteDriver {
	return &RemoteDriver{url: "http://test.local"}
}

// ---------------------------------------------------------------------------
// Status / RemoteConnectionStatus
// ---------------------------------------------------------------------------

func TestRemoteDriver_Status_Default(t *testing.T) {
	d := newTestDriver()
	if d.Status() != StatusUnknown {
		t.Errorf("initial status = %v, want StatusUnknown", d.Status())
	}
}

func TestRemoteDriver_RemoteConnectionStatus(t *testing.T) {
	d := newTestDriver()
	if got := d.RemoteConnectionStatus(); got != "unknown" {
		t.Errorf("initial = %q, want %q", got, "unknown")
	}

	d.status.Store(int32(StatusConnected))
	if got := d.RemoteConnectionStatus(); got != "connected" {
		t.Errorf("after connect = %q, want %q", got, "connected")
	}

	d.status.Store(int32(StatusDisconnected))
	if got := d.RemoteConnectionStatus(); got != "disconnected" {
		t.Errorf("after disconnect = %q, want %q", got, "disconnected")
	}
}

// ---------------------------------------------------------------------------
// trackStatus
// ---------------------------------------------------------------------------

func TestTrackStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want RemoteStatus
	}{
		{
			name: "nil error sets connected",
			err:  nil,
			want: StatusConnected,
		},
		{
			name: "ApiError sets connected (server reachable)",
			err:  &modula.ApiError{StatusCode: 404, Message: "not found"},
			want: StatusConnected,
		},
		{
			name: "wrapped ApiError sets connected",
			err:  fmt.Errorf("operation failed: %w", &modula.ApiError{StatusCode: 500, Message: "server error"}),
			want: StatusConnected,
		},
		{
			name: "network error sets disconnected",
			err:  errors.New("connection refused"),
			want: StatusDisconnected,
		},
		{
			name: "context deadline sets disconnected",
			err:  context.DeadlineExceeded,
			want: StatusDisconnected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDriver()
			d.trackStatus(tt.err)
			if got := d.Status(); got != tt.want {
				t.Errorf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isApiError / findApiError
// ---------------------------------------------------------------------------

func TestIsApiError(t *testing.T) {
	t.Run("nil error returns false", func(t *testing.T) {
		var target *modula.ApiError
		if isApiError(nil, &target) {
			t.Error("expected false for nil error")
		}
	})
	t.Run("direct ApiError returns true", func(t *testing.T) {
		var target *modula.ApiError
		err := &modula.ApiError{StatusCode: 400, Message: "bad request"}
		if !isApiError(err, &target) {
			t.Fatal("expected true for direct ApiError")
		}
		if target.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", target.StatusCode)
		}
	})
	t.Run("wrapped ApiError returns true", func(t *testing.T) {
		var target *modula.ApiError
		inner := &modula.ApiError{StatusCode: 403, Message: "forbidden"}
		err := fmt.Errorf("remote: ListRoutes: %w", inner)
		if !isApiError(err, &target) {
			t.Fatal("expected true for wrapped ApiError")
		}
		if target.StatusCode != 403 {
			t.Errorf("StatusCode = %d, want 403", target.StatusCode)
		}
	})
	t.Run("generic error returns false", func(t *testing.T) {
		var target *modula.ApiError
		if isApiError(errors.New("generic"), &target) {
			t.Error("expected false for generic error")
		}
	})
}

// ---------------------------------------------------------------------------
// doRead / doWrite / doWriteErr
// ---------------------------------------------------------------------------

func TestDoRead_Success(t *testing.T) {
	d := newTestDriver()
	result, err := doRead(d, func() (string, error) {
		return "data", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "data" {
		t.Errorf("result = %q, want %q", result, "data")
	}
	if d.Status() != StatusConnected {
		t.Errorf("status = %v, want StatusConnected", d.Status())
	}
}

func TestDoRead_NonRetryableError(t *testing.T) {
	d := newTestDriver()
	_, err := doRead(d, func() (string, error) {
		return "", &modula.ApiError{StatusCode: 404, Message: "not found"}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	// Server responded, so status should be connected
	if d.Status() != StatusConnected {
		t.Errorf("status = %v, want StatusConnected (server reachable)", d.Status())
	}
}

func TestDoRead_NetworkError(t *testing.T) {
	d := newTestDriver()
	_, err := doRead(d, func() (string, error) {
		return "", errors.New("connection refused")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if d.Status() != StatusDisconnected {
		t.Errorf("status = %v, want StatusDisconnected", d.Status())
	}
}

func TestDoWrite_Success(t *testing.T) {
	d := newTestDriver()
	result, err := doWrite(d, func() (int32, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Errorf("result = %d, want 42", result)
	}
	if d.Status() != StatusConnected {
		t.Errorf("status = %v, want StatusConnected", d.Status())
	}
}

func TestDoWrite_NoRetryOnError(t *testing.T) {
	d := newTestDriver()
	calls := 0
	_, err := doWrite(d, func() (string, error) {
		calls++
		return "", &modula.ApiError{StatusCode: 503, Message: "service unavailable"}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (writes should not retry)", calls)
	}
}

func TestDoWriteErr_Success(t *testing.T) {
	d := newTestDriver()
	err := doWriteErr(d, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status() != StatusConnected {
		t.Errorf("status = %v, want StatusConnected", d.Status())
	}
}

func TestDoWriteErr_Error(t *testing.T) {
	d := newTestDriver()
	err := doWriteErr(d, func() error {
		return errors.New("connection refused")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if d.Status() != StatusDisconnected {
		t.Errorf("status = %v, want StatusDisconnected", d.Status())
	}
}

// ---------------------------------------------------------------------------
// Unsupported / ErrRemoteMode methods
// ---------------------------------------------------------------------------

func TestUnsupportedMethods(t *testing.T) {
	d := newTestDriver()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"CreateAllTables", func() error { return d.CreateAllTables() }},
		{"CreateBootstrapData", func() error { return d.CreateBootstrapData("hash") }},
		{"CleanupBootstrapData", func() error { return d.CleanupBootstrapData() }},
		{"DropAllTables", func() error { return d.DropAllTables() }},
		{"SortTables", func() error { return d.SortTables() }},
		{"ValidateBootstrapData", func() error { return d.ValidateBootstrapData() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatal("expected ErrNotSupported")
			}
			var notSupported ErrNotSupported
			if !errors.As(err, &notSupported) {
				t.Fatalf("error type = %T, want ErrNotSupported", err)
			}
			if notSupported.Method != tt.name {
				t.Errorf("Method = %q, want %q", notSupported.Method, tt.name)
			}
		})
	}
}

func TestGetConnection_ReturnsErrRemoteMode(t *testing.T) {
	d := newTestDriver()
	conn, ctx, err := d.GetConnection()
	if conn != nil {
		t.Error("expected nil connection")
	}
	if ctx != nil {
		t.Error("expected nil context")
	}
	if !errors.Is(err, ErrRemoteMode) {
		t.Errorf("error = %v, want ErrRemoteMode", err)
	}
}

func TestExecuteQuery_ReturnsErrNotSupported(t *testing.T) {
	d := newTestDriver()
	rows, err := d.ExecuteQuery("SELECT 1", "test")
	if rows != nil {
		t.Error("expected nil rows")
	}
	var notSupported ErrNotSupported
	if !errors.As(err, &notSupported) {
		t.Fatalf("error type = %T, want ErrNotSupported", err)
	}
}

func TestQuery_ReturnsErrNotSupported(t *testing.T) {
	d := newTestDriver()
	result, err := d.Query(nil, "SELECT 1")
	if result != nil {
		t.Error("expected nil result")
	}
	var notSupported ErrNotSupported
	if !errors.As(err, &notSupported) {
		t.Fatalf("error type = %T, want ErrNotSupported", err)
	}
}

func TestGetForeignKeys_ReturnsNil(t *testing.T) {
	d := newTestDriver()
	result := d.GetForeignKeys([]string{"users"})
	if result != nil {
		t.Error("expected nil rows")
	}
}

func TestScanForeignKeyQueryRows_ReturnsNil(t *testing.T) {
	d := newTestDriver()
	result := d.ScanForeignKeyQueryRows(nil)
	if result != nil {
		t.Error("expected nil slice")
	}
}

func TestSelectColumnFromTable_NoOp(t *testing.T) {
	d := newTestDriver()
	// Should not panic
	d.SelectColumnFromTable("table", "column")
}

// ---------------------------------------------------------------------------
// Status transitions (thread-safety smoke test)
// ---------------------------------------------------------------------------

func TestStatus_ConcurrentAccess(t *testing.T) {
	d := newTestDriver()
	done := make(chan struct{})

	// Writer goroutine
	go func() {
		defer close(done)
		for i := range 100 {
			if i%2 == 0 {
				d.trackStatus(nil)
			} else {
				d.trackStatus(errors.New("network error"))
			}
		}
	}()

	// Reader goroutine (concurrent reads should not race)
	for range 100 {
		_ = d.Status()
		_ = d.RemoteConnectionStatus()
	}

	<-done

	// Final state should be one of the valid statuses
	status := d.Status()
	if status != StatusConnected && status != StatusDisconnected {
		t.Errorf("unexpected final status: %v", status)
	}
}
