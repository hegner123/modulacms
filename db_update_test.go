package main

import (
	"fmt"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestUpdateUser(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(1)
	params := mdb.UpdateUserParams{
		Datemodified: ns(times),
		Username:     ns("system"),
		Name:         ns("system"),
		Email:        ns("test@modulacms.com"),
		Hash:         ns("has"),
		Role:         ns("admin"),
		ID:           id,
	}

	updatedUser := dbUpdateUser(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Name)

	if updatedUser != expected {
		t.FailNow()
	}
}
func TestUpdateAdminRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	params := mdb.UpdateAdminRouteParams{
		Author:       ns("system"),
		Authorid:     ns("0"),
		Slug:         ns("/test"),
		Title:        ns("Test"),
		Status:       ni(0),
		Datemodified: ns(times),
		Content:      ns("Test content"),
		Template:     ns("page.html"),
		Slug_2:       ns("/test"),
	}

	updatedAdminRoute := dbUpdateAdminRoute(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Slug)

	if updatedAdminRoute != expected {
		t.FailNow()
	}
}
func TestUpdateRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	params := mdb.UpdateRouteParams{
		Author:       ns("system"),
		Authorid:     ns("0"),
		Slug:         ns("/test11"),
		Title:        ns("Test"),
		Status:       ni(0),
		Datemodified: ns(times),
		Content:      ns("Test content"),
		Template:     ns("page.html"),
		Slug_2:       ns("/test11"),
	}

	updatedRoute := dbUpdateRoute(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Slug)

	if updatedRoute != expected {
		t.FailNow()
	}
}
func TestUpdateMedia(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(1)

	params := mdb.UpdateMediaParams{
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             ns("system"),
		Authorid:           ni(0),
		Datemodified:       ns(times),
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png"),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png"),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png"),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png"),
		ID:                 id,
	}

	updatedMedia := dbUpdateMedia(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Name)

	if updatedMedia != expected {
		t.FailNow()
	}
}
func TestUpdateField(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(1)
	params := mdb.UpdateFieldParams{
		Routeid:      int64(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       ns("system"),
		Authorid:     ns("0"),
		Datecreated:  ns(times),
		Datemodified: ns(times),
		ID:           id,
	}

	updatedField := dbUpdateField(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedField != expected {
		t.FailNow()
	}
}
func TestUpdateMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

    params:=  mdb.UpdateMediaDimensionParams{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
	}

	updatedMediaDimension := dbUpdateMediaDimension(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedMediaDimension != expected {
		t.FailNow()
	}
}
func TestUpdateTables(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
    id:= int64(1)
    params:=mdb.UpdateTableParams{
        Label: ns("Tested"),
        ID: id,
    }

	updatedTable := dbUpdateTable(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedTable != expected {
		t.FailNow()
	}
}
