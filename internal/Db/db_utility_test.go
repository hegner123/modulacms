package db

import (
	"fmt"
	"testing"
)

func TestGetSchema(t *testing.T) {
	dbc := GetDb(Database{Src: getTestTable})
	m, err := GetTableColumns(dbc.Context, dbc.Connection, "datatypes")
	if err != nil {
		t.Fail()
	}

	fmt.Println(m)

}
