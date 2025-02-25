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
	_, err := DeleteUser(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
}

func TestDeleteAdminRoute(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	slug := "/to_delete"
	_, err := DeleteAdminRoute(db.Connection, db.Context, slug)
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}
}

func TestDeleteField(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	_, err := DeleteField(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}

}

func TestDeleteMedia(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	_, err := DeleteMedia(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}

}

func TestDeleteMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	id := 1
	_, err := DeleteMediaDimension(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}

}

func TestDeleteRoute(t *testing.T) {
	db := GetDb(Database{Src: deleteTestTable})
	slug := "/test1"
	_, err := DeleteRoute(db.Connection, db.Context, slug)
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}

}

func TestDeleteTables(t *testing.T) {
	id := 1
	db := GetDb(Database{Src: deleteTestTable})
	_, err := DeleteTable(db.Connection, db.Context, int64(id))
	if err != nil {
		fmt.Printf("%v", err)
		t.FailNow()
	}

}
