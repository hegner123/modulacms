package db

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var listTestTable string

func TestDBCopy(t *testing.T) {
	testTable, err := CopyDb("list_tests.db", false)
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	listTestTable = testTable
}

func TestListAdminDatatype(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.AdminDatatypes, error) {
		return ListAdminDatatypes(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListAdminField(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.AdminFields, error) {
		return ListAdminFields(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListAdminRoute(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.AdminRoutes, error) {
		return ListAdminRoute(db.Connection, db.Context)
	}()

	if err != nil {
		t.FailNow()
		return
	}
}

func TestListContentData(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.ContentData, error) {
		return ListContentData(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListContentField(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.ContentFields, error) {
		return ListContentField(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListDatatype(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Datatypes, error) {
		return ListDatatype(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListDatatypeByRoute(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.ListDatatypeByRouteIdRow, error) {
		return ListDatatypeById(db.Connection, db.Context, 1)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListField(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Fields, error) {
		return ListField(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListFieldByRoute(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.ListFieldByRouteIdRow, error) {
		return ListFieldByRouteId(db.Connection, db.Context, 1)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListMedia(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Media, error) {
		return ListMedia(db.Connection, db.Context)
	}()

	if err != nil {
		t.FailNow()
		return
	}
}

func TestListMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.MediaDimensions, error) {
		return ListMediaDimension(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListRoles(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Roles, error) {
		return ListRoles(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListRoute(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Routes, error) {
		return ListRoute(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListChildrenOfRoute(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	datas, err := ListDatatypeById(db.Connection, db.Context, 1)
	if err != nil {
		return
	}

	field, err := ListFieldByRouteId(db.Connection, db.Context, 1)
	if err != nil {
		return
	}

	file, err := os.Create("log.txt")
	if err != nil {
		fmt.Printf("%v", err)
	}
	dataMap := map[string][]mdb.ListDatatypeByRouteIdRow{
		"Datatypes": *datas,
	}
	fieldMap := map[string][]mdb.ListFieldByRouteIdRow{
		"Fields": *field,
	}
	w := json.NewEncoder(file)
	err = w.Encode(dataMap)
	if err != nil {
		fmt.Printf("%v", err)
	}
	err = w.Encode(fieldMap)
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func TestListTables(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Tables, error) {
		return ListTable(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListTokens(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Tokens, error) {
		return ListTokens(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}

func TestListUser(t *testing.T) {
	db := GetDb(Database{Src: listTestTable})
	_, err := func() (*[]mdb.Users, error) {
		return ListUser(db.Connection, db.Context)
	}()
	if err != nil {
		t.FailNow()
		return
	}
}
