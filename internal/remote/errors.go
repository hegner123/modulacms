package remote

import "errors"

// ErrRemoteMode is returned by methods that require a local database connection
// (GetConnection, DropAllTables, CreateAllTables, etc.).
var ErrRemoteMode = errors.New("operation not available in remote mode")

// ErrNotSupported is returned by DbDriver methods that have no remote equivalent.
// The error message includes the method name for debuggability.
type ErrNotSupported struct {
	Method string
}

func (e ErrNotSupported) Error() string {
	return "remote: method not supported: " + e.Method
}
