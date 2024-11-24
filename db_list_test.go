package main

import (
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestListUser(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListUser(db, ctx)
	}()

	if _, ok := res.([]mdb.User); ok {
		return
	} else {
		t.FailNow()
	}
}
func TestListAdminRoute(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListAdminRoute(db, ctx)
	}()

	if _, ok := res.([]mdb.Adminroute); ok {
		return
	} else {
		t.FailNow()
	}
}
func TestListRoute(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListRoute(db, ctx)
	}()

	if _, ok := res.([]mdb.Route); ok {
		return
	} else {
		t.FailNow()
	}
}
func TestListMedia(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
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
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListField(db, ctx)
	}()

	if _, ok := res.([]mdb.Field); ok {
		return
	} else {
		t.FailNow()
	}
}
func TestListMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	res := func() interface{} {
		return dbListMediaDimension(db, ctx)
	}()

	if _, ok := res.([]mdb.MediaDimension); ok {
		return
	} else {
		t.FailNow()
	}
}
func TestListTables(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
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
