package remote

import (
	"context"
	"errors"
	"net"
	"time"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// retryRead wraps a read operation with a single retry on transient failures.
// Uses Go generics so it works with any return type (*[]db.Routes, *db.ContentData, etc.).
// Retry rules: 1s delay, maximum 1 retry (2 total attempts).
func retryRead[T any](fn func() (T, error)) (T, error) {
	result, err := fn()
	if err == nil || !isRetryable(err) {
		return result, err
	}
	time.Sleep(1 * time.Second)
	return fn()
}

// isRetryable checks if an error is a transient failure worth retrying.
// Retries on: timeout, 502 Bad Gateway, 503 Service Unavailable, 504 Gateway Timeout.
// No retry on: 4xx (client error), connection refused (server down), unknown errors.
func isRetryable(err error) bool {
	// Check for Go SDK API errors with retryable status codes
	var apiErr *modula.ApiError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 502, 503, 504:
			return true
		}
		return false
	}
	// Check for network timeouts (net.Error interface)
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	// Check for context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}
