package email

import "fmt"

// InvalidMessageError indicates a message validation failure.
// Callers inspect Field and Problem to determine which part of the message
// failed validation.
type InvalidMessageError struct {
	Field   string
	Problem string
}

func (e *InvalidMessageError) Error() string {
	return fmt.Sprintf("invalid email message: %s: %s", e.Field, e.Problem)
}

// ProviderError indicates the remote email provider rejected the request.
// Code holds the HTTP status code (0 for SMTP/dial errors). Unwrap returns
// the underlying provider error.
type ProviderError struct {
	Provider string
	Code     int
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("email provider %s returned %d: %v", e.Provider, e.Code, e.Err)
	}
	return fmt.Sprintf("email provider %s: %v", e.Provider, e.Err)
}

func (e *ProviderError) Unwrap() error { return e.Err }

// DisabledError indicates email sending is disabled in the configuration.
type DisabledError struct{}

func (e *DisabledError) Error() string {
	return "email sending is disabled"
}
