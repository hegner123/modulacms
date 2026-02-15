package utility

import (
	"database/sql"
)

// HandleRowsCloseDeferErr closes the database rows and logs any error.
func HandleRowsCloseDeferErr(r *sql.Rows) {
	if err := r.Close(); err != nil {
		DefaultLogger.Warn("Failed to close database rows", err)
	}
}

// HandleConnectionCloseDeferErr closes the database connection and logs any error.
func HandleConnectionCloseDeferErr(r *sql.DB) {
	if err := r.Close(); err != nil {
		DefaultLogger.Warn("Failed to close database connection", err)
	}
}
