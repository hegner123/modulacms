package main

import (
	"encoding/json"
	"os"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var listTestTable string
func TestDBCopy(t *testing.T) {
	testTable,err := createDbCopy("list_tests.db",false)
    if err != nil { 
        logError("failed to create copy of the database, I have to hurry, I'm running out of time!!! ", err)
        t.FailNow()
    }
	listTestTable = testTable
}

func TestListUser(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListUser(db, ctx)
	}()

	if _, ok := res.([]mdb.Users); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListAdminRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListAdminRoute(db, ctx)
	}()

	if _, ok := res.([]mdb.AdminRoutes); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListRoute(db, ctx)
	}()

	if _, ok := res.([]mdb.Routes); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListMedia(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListMedia(db, ctx)
	}()

	if _, ok := res.([]mdb.Media); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListField(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListField(db, ctx)
	}()

	if _, ok := res.([]mdb.Fields); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListMediaDimension(db, ctx)
	}()

	if _, ok := res.([]mdb.MediaDimensions); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListTables(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListTable(db, ctx)
	}()

	if _, ok := res.([]mdb.Tables); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListDatatype(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListDatatype(db, ctx)
	}()

	if _, ok := res.([]mdb.Datatypes); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListDatatypeByRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListDatatypeById(db, ctx, 1)
	}()

	if _, ok := res.([]mdb.ListDatatypeByRouteIdRow); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListFieldByRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListFieldById(db, ctx, 1)
	}()

	if _, ok := res.([]mdb.ListFieldByRouteIdRow); ok {
		return
	} else {
		t.FailNow()
	}
}

func TestListChildrenOfRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: listTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	datas := dbListDatatypeById(db, ctx, 1)

	field := dbListFieldById(db, ctx, 1)

	file, err := os.Create("log.txt")
	if err != nil {
		logError("failed to create file ", err)
	}
	dataMap := map[string][]mdb.ListDatatypeByRouteIdRow{
		"Datatypes": datas,
	}
	fieldMap := map[string][]mdb.ListFieldByRouteIdRow{
		"Fields": field,
	}
	w := json.NewEncoder(file)
	err = w.Encode(dataMap)
	if err != nil {
		logError("failed to encode datas", err)
	}
	err = w.Encode(fieldMap)
	if err != nil {
		logError("failed to encode field", err)
	}
}
