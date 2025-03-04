package db

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)


var psql_c config.Config = config.Config{
	Db_Driver:   "postgres",
	Db_Name:     "modula_db",
	Db_URL:      "localhost",
	Db_User:     "modula_u",
	Db_Password: "modula_pass",
}

func TestForeignKeySetupPsql(t *testing.T) {
	d := ConfigDB(psql_c)
	err := d.CreateAllTables()
	if err != nil {
		t.Fatal(err)
	}
}
