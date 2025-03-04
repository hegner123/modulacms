package db

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

var mysql_c config.Config = config.Config{
	Db_Driver:   "mysql",
	Db_Name:     "modula_db",
	Db_URL:      "localhost",
	Db_User:     "modula_u",
	Db_Password: "modula_pass",
}

func TestForeignKeySetup(t *testing.T) {
	d := ConfigDB(mysql_c)
	err := d.CreateAllTables()
	if err != nil {
		t.Fatal(err)
	}
}
/*
func TestCreateAdminDatatype(t *testing.T) {
	d := ConfigDB(mysql_c)
	params := CreateAdminDatatypeParams{
		AdminRouteID: Ni(1),
		ParentID:     sql.NullInt64{Valid: false},
		Label:        "Test",
		Type:         "Test",
		Author:       "system",
		AuthorID:     int64(1),
	}
	row := d.CreateAdminDatatype(params)
	fmt.Println(row)

}
*/
