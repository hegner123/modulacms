package development

import "github.com/hegner123/modulacms/internal/db"

func HandleDatabaseCloseDeferErr(d *db.Database) {
	err := d.Connection.Close()
	if err != nil {
		return
	}

}
