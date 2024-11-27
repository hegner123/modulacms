package main

import (
	"fmt"
	"reflect"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestInit(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect to database ", err)
	}
	createSetupInserts(db, ctx, "1")
}

func TestGetUser(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(2)

	userRow, err := dbGetUser(db, ctx, id)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	expected := mdb.User{
		ID:           int64(2),
		Datecreated:  userRow.Datecreated,
		Datemodified: userRow.Datemodified,
		Username:     "system",
		Name:         "system",
		Email:        "system@modulacms.com1",
		Hash:         "has",
		Role:         "admin",
	}

	if reflect.DeepEqual(userRow, expected) {
		t.FailNow()
	}
}

/*
	func TestGetUserId(t *testing.T){
		db, ctx, err := getDb(Database{DB: "modula_test.db"})
		if err != nil {
			logError("failed to connect or create database", err)
		}
		defer db.Close()
	    id := dbGetUserId(db,ctx,1)
	}
*/
func TestGetAdminRoute(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	adminRouteRow := dbGetAdminRoute(db, ctx, "/test11")

	expected := mdb.Adminroute{
		ID:           int64(1),
		Author:       ns("system"),
		Authorid:     1,
		Slug:         "/test11",
		Title:        "Test",
		Status:       int64(0),
		Datecreated:  adminRouteRow.Datecreated,
		Datemodified: adminRouteRow.Datemodified,
		Template:     ns("page.html"),
	}

	if reflect.DeepEqual(adminRouteRow, expected) {
		t.FailNow()
	}
}

func TestGetRoute(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	routeRow := dbGetRoute(db, ctx, "/test11")

	expected := mdb.Route{
		ID:           int64(1),
		Author:       "system",
		Authorid:     1,
		Slug:         "/test11",
		Title:        "Test",
		Status:       int64(0),
		Datecreated:  routeRow.Datecreated,
		Datemodified: routeRow.Datemodified,
	}

	if reflect.DeepEqual(routeRow, expected) {
		t.FailNow()
	}
}

func TestGetMedia(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(2)

	mediaRow := dbGetMedia(db, ctx, id)

	expected := mdb.Media{
		ID:                 int64(1),
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "system",
		Authorid:           1,
		Datecreated:        mediaRow.Datecreated,
		Datemodified:       mediaRow.Datemodified,
		Url:                ns("public/2024/11/test.png1"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png"),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png"),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png"),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png"),
	}

	if reflect.DeepEqual(mediaRow, expected) {
		t.FailNow()
	}
}

func TestGetField(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(3)
	fieldRow := dbGetField(db, ctx, id)

	expected := mdb.Field{
		Routeid:      ni(1),
		Parentid:     ni(1),
		Label:        "title",
		Data:         "Test Field",
		Type:         "text",
		Struct:       ns("text"),
		Author:       ns("system"),
		Authorid:     1,
		Datecreated:  fieldRow.Datecreated,
		Datemodified: fieldRow.Datemodified,
	}

	if reflect.DeepEqual(fieldRow, expected) {
		t.FailNow()
	}
}

func TestGetMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(2)

	mediaDimensionRow := dbGetMediaDimension(db, ctx, id)

	expected := mdb.MediaDimension{
		Label:  ns("Desktop1"),
		Width:  ni(1920),
		Height: ni(1080),
	}

	if reflect.DeepEqual(mediaDimensionRow, expected) {
		t.FailNow()
	}
}

func TestGetTables(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(2)
	tableRow := dbGetTable(db, ctx, id)

	expected := mdb.Tables{
		Label: ns("Test1"),
	}

	if reflect.DeepEqual(tableRow, expected) {
		t.FailNow()
	}
}
