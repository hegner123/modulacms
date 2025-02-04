package cli

import (
	"fmt"
	"testing"

	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var tableName string = "fields"

func TestCliConnect(t *testing.T) {
	h := fmt.Sprintf("Field for table %s", tableName)
	utility.LogHeader(h)


}

func TestForeignKeys(t *testing.T) {
	h := fmt.Sprintf("ForeignKeys for table %s\n", tableName)
	utility.LogHeader(h)
	dbc := db.GetDb(db.Database{Src: "get_tests.db"})
	GetRelationships(tableName, dbc)

}
