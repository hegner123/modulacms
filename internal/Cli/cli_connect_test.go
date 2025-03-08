package cli

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var tableName string = "fields"

func TestCliConnect(t *testing.T) {
	h := fmt.Sprintf("Field for table %s", tableName)
	utility.LogHeader(h)

}

func TestForeignKeys(t *testing.T) {
	c := config.Config{
		Db_Driver: "sqlite",
		Db_Name:   "get_tests.db",
	}
	h := fmt.Sprintf("ForeignKeys for table %s\n", tableName)
	utility.LogHeader(h)
	dbc := db.ConfigDB(c)
	GetRelationships(tableName, dbc.(db.Database))

}
