package service

import (
	"errors"
	"fmt"
	"strings"
)

// NotFoundError indicates the requested resource does not exist.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
}

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field   string
	Message string
}

// ValidationError holds one or more field-level validation failures.
// Services collect all errors before returning, rather than failing on the first.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("validation: %s: %s", e.Errors[0].Field, e.Errors[0].Message)
	}
	parts := make([]string, len(e.Errors))
	for i, fe := range e.Errors {
		parts[i] = fe.Field + ": " + fe.Message
	}
	return fmt.Sprintf("validation: %s", strings.Join(parts, "; "))
}

// HasErrors returns true if any field errors have been collected.
func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

// NewValidationError creates a ValidationError with a single field error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Errors: []FieldError{{Field: field, Message: message}},
	}
}

// NewValidationErrors creates a ValidationError from multiple field errors.
func NewValidationErrors(errs ...FieldError) *ValidationError {
	return &ValidationError{Errors: errs}
}

// Add appends a field error and returns the receiver for chaining.
func (e *ValidationError) Add(field, message string) *ValidationError {
	e.Errors = append(e.Errors, FieldError{Field: field, Message: message})
	return e
}

// ConflictError indicates a resource already exists or a uniqueness constraint was violated.
type ConflictError struct {
	Resource string
	ID       string
	Detail   string
}

func (e *ConflictError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s %q conflict: %s", e.Resource, e.ID, e.Detail)
	}
	return fmt.Sprintf("%s %q already exists", e.Resource, e.ID)
}

// ForbiddenError indicates the caller lacks permission for the requested operation.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("forbidden: %s", e.Message)
	}
	return "forbidden"
}

// InternalError wraps an unexpected error from a lower layer.
type InternalError struct {
	Err error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error: %s", e.Err)
}

func (e *InternalError) Unwrap() error {
	return e.Err
}

// IsNotFound reports whether err is a *NotFoundError.
func IsNotFound(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// IsValidation reports whether err is a *ValidationError.
func IsValidation(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}

// IsConflict reports whether err is a *ConflictError.
func IsConflict(err error) bool {
	var target *ConflictError
	return errors.As(err, &target)
}

// IsForbidden reports whether err is a *ForbiddenError.
func IsForbidden(err error) bool {
	var target *ForbiddenError
	return errors.As(err, &target)
}
