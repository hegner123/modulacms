package db

import (
	"fmt"
	"testing"
)

var deleteTestTable string

func TestDeleteDBCopy(t *testing.T) {
	testTable, err := createDbCopy("delete_tests.db",false)
	if err != nil {
		logError("failed to create copy of the database, I have to hurry, I'm running out of time!!! ", err)
		t.FailNow()
	}
	deleteTestTable = testTable
}

func TestDeleteUser(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := 2
	result := dbDeleteUser(db, ctx, int64(id))
	expected := fmt.Sprintf("Deleted User %d successfully", id)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteAdminRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	slug := "/to_delete"
	result := dbDeleteAdminRoute(db, ctx, slug)
	expected := fmt.Sprintf("Deleted Admin Route %s successfully", slug)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteField(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := 1
	result := dbDeleteField(db, ctx, int64(id))
	expected := fmt.Sprintf("Deleted Field %d successfully", id)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteMedia(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := 1
	result := dbDeleteMedia(db, ctx, int64(id))
	expected := fmt.Sprintf("Deleted Media %d successfully", id)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := 1
	result := dbDeleteMediaDimension(db, ctx, int64(id))
	expected := fmt.Sprintf("Deleted Media Dimension %d successfully", id)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteRoute(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	slug := "/test1"
	result := dbDeleteRoute(db, ctx, slug)
	expected := fmt.Sprintf("Deleted Route %s successfully", slug)
	if expected != result {
		t.FailNow()
	}
}

func TestDeleteTables(t *testing.T) {
	db, ctx, err := getDb(Database{src: deleteTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := 1
	result := dbDeleteTable(db, ctx, int64(id))
	expected := fmt.Sprintf("Deleted Table %d successfully", id)
	if expected != result {
		t.FailNow()
	}
}
