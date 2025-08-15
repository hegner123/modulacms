package utility

import (
	"database/sql"
)

func HandleRowsCloseDeferErr(r *sql.Rows) {
	err := r.Close()
	if err != nil {
		return
	}

}
func HandleConnectionCloseDeferErr(r *sql.DB) {
	err := r.Close()
	if err != nil {
		return
	}

}
