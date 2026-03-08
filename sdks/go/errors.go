package modula

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ApiError represents a non-2xx HTTP response from the ModulaCMS API.
//
// Every failed API call returns an *ApiError, which can be unwrapped with
// [errors.As] or checked with the convenience functions [IsNotFound],
// [IsUnauthorized], [IsDuplicateMedia], and [IsInvalidMediaPath].
//
// StatusCode is the HTTP status code (e.g. 404, 401, 500). Message is a
// human-readable error extracted from the JSON response body when available.
// Body is the raw response body string, useful for debugging when Message
// is empty.
type ApiError struct {
	// StatusCode is the HTTP response status code.
	StatusCode int

	// Message is the server-provided error message, extracted from a JSON
	// "message" or "error" field in the response body. May be empty if the
	// response is not JSON or lacks these fields.
	Message string

	// Body is the full response body as a string. Useful for debugging
	// when Message is empty or for logging the exact server response.
	Body string
}

// Error returns a formatted error string including the status code and
// either the server message or the standard HTTP status text.
func (e *ApiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("modula: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("modula: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// IsNotFound reports whether err is (or wraps) an [*ApiError] with HTTP
// status 404. Use this after [Resource.Get] or [Resource.Delete] to
// distinguish "not found" from other failures.
func IsNotFound(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized reports whether err is (or wraps) an [*ApiError] with
// HTTP status 401. This typically indicates a missing, expired, or invalid
// API key.
func IsUnauthorized(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsDuplicateMedia reports whether err is (or wraps) an [*ApiError] with
// HTTP status 409 (Conflict), indicating that a media upload was rejected
// because a file with the same content hash already exists.
func IsDuplicateMedia(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsInvalidMediaPath reports whether err is (or wraps) an [*ApiError] with
// HTTP status 400 whose body mentions path traversal or invalid path
// characters. This occurs when a media upload path contains ".." segments
// or disallowed characters.
func IsInvalidMediaPath(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusBadRequest &&
			(strings.Contains(apiErr.Body, "path traversal") || strings.Contains(apiErr.Body, "invalid character in path"))
	}
	return false
}
