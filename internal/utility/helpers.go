package utility

import (
	"database/sql"
)

func HandleRowsCloseDeferErr(r *sql.Rows) {
	if err := r.Close(); err != nil {
		DefaultLogger.Warn("Failed to close database rows", err)
	}
}

func HandleConnectionCloseDeferErr(r *sql.DB) {
	if err := r.Close(); err != nil {
		DefaultLogger.Warn("Failed to close database connection", err)
	}
}
