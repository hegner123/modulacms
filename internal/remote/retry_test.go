package remote

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "502 bad gateway",
			err:  &modula.ApiError{StatusCode: 502, Message: "bad gateway"},
			want: true,
		},
		{
			name: "503 service unavailable",
			err:  &modula.ApiError{StatusCode: 503, Message: "service unavailable"},
			want: true,
		},
		{
			name: "504 gateway timeout",
			err:  &modula.ApiError{StatusCode: 504, Message: "gateway timeout"},
			want: true,
		},
		{
			name: "400 bad request not retryable",
			err:  &modula.ApiError{StatusCode: 400, Message: "bad request"},
			want: false,
		},
		{
			name: "401 unauthorized not retryable",
			err:  &modula.ApiError{StatusCode: 401, Message: "unauthorized"},
			want: false,
		},
		{
			name: "404 not found not retryable",
			err:  &modula.ApiError{StatusCode: 404, Message: "not found"},
			want: false,
		},
		{
			name: "500 internal server error not retryable",
			err:  &modula.ApiError{StatusCode: 500, Message: "internal server error"},
			want: false,
		},
		{
			name: "network timeout is retryable",
			err:  &timeoutError{timeout: true},
			want: true,
		},
		{
			name: "network error without timeout not retryable",
			err:  &timeoutError{timeout: false},
			want: false,
		},
		{
			name: "context deadline exceeded is retryable",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "context canceled not retryable",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "generic error not retryable",
			err:  errors.New("something failed"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryRead_SuccessOnFirst(t *testing.T) {
	calls := 0
	result, err := retryRead(func() (string, error) {
		calls++
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q, want %q", result, "ok")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetryRead_NonRetryableError(t *testing.T) {
	calls := 0
	_, err := retryRead(func() (string, error) {
		calls++
		return "", &modula.ApiError{StatusCode: 404, Message: "not found"}
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (should not retry on 404)", calls)
	}
}

func TestRetryRead_RetryableErrorThenSuccess(t *testing.T) {
	calls := 0
	result, err := retryRead(func() (string, error) {
		calls++
		if calls == 1 {
			return "", &modula.ApiError{StatusCode: 503, Message: "service unavailable"}
		}
		return "recovered", nil
	})
	if err != nil {
		t.Fatalf("unexpected error on retry: %v", err)
	}
	if result != "recovered" {
		t.Errorf("result = %q, want %q", result, "recovered")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestRetryRead_RetryableErrorThenError(t *testing.T) {
	calls := 0
	_, err := retryRead(func() (string, error) {
		calls++
		return "", &modula.ApiError{StatusCode: 503, Message: "service unavailable"}
	})
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
}

func TestRetryRead_GenericTypeSupport(t *testing.T) {
	type payload struct {
		Value int32
	}
	result, err := retryRead(func() (*payload, error) {
		return &payload{Value: 42}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.Value != 42 {
		t.Errorf("result = %+v, want &{Value:42}", result)
	}
}

// timeoutError implements net.Error for testing timeout detection.
type timeoutError struct {
	timeout bool
}

func (e *timeoutError) Error() string   { return "test timeout error" }
func (e *timeoutError) Timeout() bool   { return e.timeout }
func (e *timeoutError) Temporary() bool { return false }

// Compile-time check that timeoutError satisfies net.Error.
var _ net.Error = (*timeoutError)(nil)

// Ensure the 1-second retry delay is real (not accidentally mocked).
func TestRetryRead_DelayOnRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping delay test in short mode")
	}
	start := time.Now()
	calls := 0
	_, _ = retryRead(func() (string, error) {
		calls++
		if calls == 1 {
			return "", context.DeadlineExceeded
		}
		return "ok", nil
	})
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Errorf("retry delay too short: %v, expected >= 1s", elapsed)
	}
}
