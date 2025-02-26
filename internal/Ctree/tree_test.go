package ctree
import (
	"fmt"
	"testing"

	db "github.com/hegner123/modulacms/internal/Db"
)

var TreeTestTable string

func TestTreeDBCopy(t *testing.T) {
	testTable, err := db.CopyDb("list-tests.db", true)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	TreeTestTable = testTable
}

func TestTreeParse(t *testing.T) {
	//tree := Tree{fn: Scan, Root: &Node{}}

	dbc := db.GetDb(db.Database{Src: TreeTestTable})
	if dbc.Err != nil {
		t.Fatal(dbc.Err)
	}
    adts, err := db.ListAdminDatatypeByAdminRouteId(dbc.Connection, dbc.Context, int64(1))
	if err != nil {
		t.Fatal(err)
	}
    a := *adts
    for i:=0;i<len(a);i++{
        v:=a[i]
        fmt.Println(v)

    }

}
