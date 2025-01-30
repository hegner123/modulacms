package db

import (
	"fmt"
	"testing"
)

var deleteTestTable string

func TestDeleteDBCopy(t *testing.T) {
	testTable, err := CopyDb("delete_tests.db", false)
	if err != nil {
		fmt.Printf("failed to create copy of the database, I have to hurry, I'm running out of time!!!  %v", err)
		t.FailNow()
	}
	deleteTestTable = testTable
}

func TestDeleteUser(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 2
	result, err := DeleteUser(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteAdminRoute(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	slug := "/to_delete"
	result, err := DeleteAdminRoute(db.Connection, db.Context, slug)
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteField(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	result, err := DeleteField(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteMedia(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	result, err := DeleteMedia(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	result, err := DeleteMediaDimension(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteRoute(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	slug := "/test1"
	result, err := DeleteRoute(db.Connection, db.Context, slug)
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}

func TestDeleteTables(t *testing.T) {
	id := 1
	db := GetDb(Database{Src: deleteTestTable})
	result, err := DeleteTable(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
	fmt.Printf("%v", result)
}
