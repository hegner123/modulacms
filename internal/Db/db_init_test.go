package db

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestMysqlConnection(t *testing.T) {
	c := config.Config{
		Db_Driver:   "mysql",
        Db_Name:     "modula_db",
		Db_URL:      "localhost",
        Db_User:     "modula_u",
		Db_Password: "modula_pass",
	}
	d := ConfigDB(c)
	err := d.Ping()
	if err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("Connected Successfully\n")
	}
    connection,_ := d.GetConnection()
    defer connection.Close()
}

func TestPsqlConnection(t *testing.T) {
	c := config.Config{
		Db_Driver:   "postgres",
        Db_Name:     "modula_db",
		Db_URL:      "localhost",
        Db_User:     "modula_u",
		Db_Password: "modula_pass",
	}
	d := ConfigDB(c)
	err := d.Ping()
	if err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("Connected Successfully\n")
	}
    connection,_ := d.GetConnection()
    defer connection.Close()
}
