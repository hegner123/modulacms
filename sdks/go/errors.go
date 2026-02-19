package modula

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ApiError represents an error response from the Modula API.
type ApiError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ApiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("modula: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("modula: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// IsNotFound reports whether the error is a 404 Not Found response.
func IsNotFound(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsUnauthorized reports whether the error is a 401 Unauthorized response.
func IsUnauthorized(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsDuplicateMedia reports whether the error is a 409 Conflict for a duplicate media upload.
func IsDuplicateMedia(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsInvalidMediaPath reports whether the error is a 400 Bad Request for an invalid media path.
func IsInvalidMediaPath(err error) bool {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusBadRequest &&
			(strings.Contains(apiErr.Body, "path traversal") || strings.Contains(apiErr.Body, "invalid character in path"))
	}
	return false
}
