package cli

// Logger defines the logging interface for the CLI package.
// This allows dependency injection of loggers for better testability
// and flexibility in logging implementations.
type Logger interface {
	// Console logging methods
	Info(message string, args ...any)
	Error(message string, err error, args ...any)
	Fatal(message string, err error, args ...any)

	// File logging methods
	Finfo(message string, args ...any)
	Ferror(message string, err error, args ...any)
	Ffatal(message string, err error, args ...any)
}

// NopLogger is a no-op logger that discards all log messages.
// Useful for testing or when logging should be disabled.
type NopLogger struct{}

func (NopLogger) Info(message string, args ...any)              {}
func (NopLogger) Error(message string, err error, args ...any)  {}
func (NopLogger) Fatal(message string, err error, args ...any)  {}
func (NopLogger) Finfo(message string, args ...any)             {}
func (NopLogger) Ferror(message string, err error, args ...any) {}
func (NopLogger) Ffatal(message string, err error, args ...any) {}
