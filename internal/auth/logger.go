package auth

// Logger is the logging interface consumed by the auth package. Callers pass
// a concrete logger (e.g. *utility.Logger) into NewTokenRefresher and
// NewUserProvisioner via their constructors.
type Logger interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Warn(message string, err error, args ...any)
	Error(message string, err error, args ...any)
}
