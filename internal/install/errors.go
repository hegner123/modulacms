package install

import "fmt"

// InstallError provides structured error information with context and user hints
type InstallError struct {
	Operation string
	Cause     error
	Hint      string
}

// Error returns the formatted error message including operation, cause, and hint if available.
func (e *InstallError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %v\nHint: %s", e.Operation, e.Cause, e.Hint)
	}
	return fmt.Sprintf("%s: %v", e.Operation, e.Cause)
}

// Unwrap returns the underlying error cause for error chain inspection.
func (e *InstallError) Unwrap() error {
	return e.Cause
}

// ErrConfigWrite creates an error for config file write failures
func ErrConfigWrite(cause error, path string) *InstallError {
	return &InstallError{
		Operation: "write config file",
		Cause:     cause,
		Hint:      fmt.Sprintf("Check that you have write permission to %q and the directory exists", path),
	}
}

// ErrDBConnect creates an error for database connection failures
func ErrDBConnect(cause error, driver string) *InstallError {
	hint := "Check database credentials and ensure the database server is running"
	if driver == "sqlite" {
		hint = "Check that the database path is writable and has a .db extension"
	}
	return &InstallError{
		Operation: "connect to database",
		Cause:     cause,
		Hint:      hint,
	}
}

// ErrDBTables creates an error for table creation failures
func ErrDBTables(cause error) *InstallError {
	return &InstallError{
		Operation: "create database tables",
		Cause:     cause,
		Hint:      "Check database permissions. For MySQL/PostgreSQL, ensure the user has CREATE TABLE privileges",
	}
}

// ErrDBBootstrap creates an error for bootstrap data insertion failures
func ErrDBBootstrap(cause error) *InstallError {
	return &InstallError{
		Operation: "insert bootstrap data",
		Cause:     cause,
		Hint:      "The database tables may already contain data. Consider using a fresh database",
	}
}

// ErrBucketConnect creates an error for S3 bucket connection failures
func ErrBucketConnect(cause error) *InstallError {
	return &InstallError{
		Operation: "connect to S3 bucket",
		Cause:     cause,
		Hint:      "Verify your access key, secret key, and endpoint URL. Check that the bucket exists and is accessible",
	}
}

// ErrValidation creates an error for input validation failures
func ErrValidation(field string, cause error) *InstallError {
	return &InstallError{
		Operation: fmt.Sprintf("validate %s", field),
		Cause:     cause,
		Hint:      "",
	}
}

// ErrMaxRetries creates an error when max install retries exceeded
func ErrMaxRetries(attempts int) *InstallError {
	return &InstallError{
		Operation: "complete installation",
		Cause:     fmt.Errorf("exceeded maximum retry attempts (%d)", attempts),
		Hint:      "Review the errors above and fix the configuration before running --install again",
	}
}

// ErrUserAborted creates an error when user cancels the installation
func ErrUserAborted() *InstallError {
	return &InstallError{
		Operation: "installation",
		Cause:     fmt.Errorf("cancelled by user"),
		Hint:      "",
	}
}
