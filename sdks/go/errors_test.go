package modula

import (
	"errors"
	"fmt"
	"testing"
)

func TestApiError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ApiError
		want string
	}{
		{
			name: "with message",
			err:  &ApiError{StatusCode: 400, Message: "invalid input"},
			want: "modula: 400 invalid input",
		},
		{
			name: "without message uses status text",
			err:  &ApiError{StatusCode: 404},
			want: "modula: 404 Not Found",
		},
		{
			name: "500 without message",
			err:  &ApiError{StatusCode: 500},
			want: "modula: 500 Internal Server Error",
		},
		{
			name: "401 with custom message",
			err:  &ApiError{StatusCode: 401, Message: "token expired"},
			want: "modula: 401 token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "404 ApiError",
			err:  &ApiError{StatusCode: 404},
			want: true,
		},
		{
			name: "404 ApiError with message",
			err:  &ApiError{StatusCode: 404, Message: "resource not found"},
			want: true,
		},
		{
			name: "400 ApiError",
			err:  &ApiError{StatusCode: 400},
			want: false,
		},
		{
			name: "500 ApiError",
			err:  &ApiError{StatusCode: 500},
			want: false,
		},
		{
			name: "wrapped 404 ApiError",
			err:  fmt.Errorf("outer: %w", &ApiError{StatusCode: 404}),
			want: true,
		},
		{
			name: "non-ApiError",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "401 ApiError",
			err:  &ApiError{StatusCode: 401},
			want: true,
		},
		{
			name: "401 ApiError with message",
			err:  &ApiError{StatusCode: 401, Message: "invalid credentials"},
			want: true,
		},
		{
			name: "403 ApiError",
			err:  &ApiError{StatusCode: 403},
			want: false,
		},
		{
			name: "404 ApiError",
			err:  &ApiError{StatusCode: 404},
			want: false,
		},
		{
			name: "wrapped 401 ApiError",
			err:  fmt.Errorf("outer: %w", &ApiError{StatusCode: 401}),
			want: true,
		},
		{
			name: "non-ApiError",
			err:  errors.New("unauthorized"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnauthorized(tt.err)
			if got != tt.want {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}
