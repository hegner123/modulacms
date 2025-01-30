package mTemplate

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

