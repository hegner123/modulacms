package db

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestDbSqliteDump(t *testing.T) {
	v := false
	c := config.LoadConfig(&v, "")
	d := ConfigDB(c)
	err := d.DumpSql(c)
	if err != nil {
		t.Fatal(err)
	}

}

func TestDbMysqlDump(t *testing.T) {
	v := false
	c := config.LoadConfig(&v, "config_mysql.json")
	d := ConfigDB(c)
	err := d.DumpSql(c)
	if err != nil {
		t.Fatal(err)
	}

}
