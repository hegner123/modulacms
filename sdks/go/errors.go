package modulacms

import (
	"errors"
	"fmt"
	"net/http"
)

// ApiError represents an error response from the ModulaCMS API.
type ApiError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ApiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("modulacms: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("modulacms: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
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
