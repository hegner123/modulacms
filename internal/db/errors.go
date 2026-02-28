package db

import "errors"

// ErrRevisionConflict indicates that a content update failed due to an
// optimistic locking conflict. Another user or process modified the content
// between the time it was read and the update was attempted.
var ErrRevisionConflict = errors.New("revision conflict: content was modified concurrently")
