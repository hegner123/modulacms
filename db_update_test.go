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
	id := int64(2)
	params := mdb.UpdateUserParams{
		Datemodified: times,
		Name:         "systemupdate",
		Hash:         "has",
		Role:         "admin",
		UserID:       id,
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
		Authorid:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datemodified: times,
		Template:     ns("page.html"),
		Slug_2:       "/test",
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
		Author:       "system",
		Authorid:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datemodified: times,
		Slug_2:       "/test",
	}

	updatedRoute := dbUpdateRoute(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Slug)

	if updatedRoute != expected {
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
	id := int64(3)
	params := mdb.UpdateFieldParams{
		Routeid:      int64(1),
		Parentid:     int64(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "system",
		Authorid:     1,
		Datemodified: ns(times),
		Datecreated:  ns(times),
		FieldID:      id,
	}

	updatedField := dbUpdateField(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedField != expected {
		t.FailNow()
	}
}

func TestUpdateDatatype(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	id := int64(1)
	params := mdb.UpdateDatatypeParams{
		Routeid:      int64(1),
		Parentid:     ni(1),
		Label:        "Parent",
		Type:         "text",
		Author:       "system",
		Authorid:     1,
		Datemodified: ns(times),
		DatatypeID:   id,
	}

	updatedDatatype := dbUpdateDatatype(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedDatatype != expected {
		t.FailNow()
	}
}

func TestUpdateMedia(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	params := mdb.UpdateMediaParams{
		Name: ns("Best"),
        Author: "system",
        Authorid: int64(1),
		ID:   int64(2),
	}

	updatedMedia := dbUpdateMedia(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Name)

	if updatedMedia != expected {
		t.FailNow()
	}
}

func TestUpdateMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	params := mdb.UpdateMediaDimensionParams{
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
	id := int64(1)
	params := mdb.UpdateTableParams{
		Label: ns("Tested"),
		ID:    id,
	}

	updatedTable := dbUpdateTable(db, ctx, params)
	expected := fmt.Sprintf("Successfully updated %v\n", params.Label)

	if updatedTable != expected {
		t.FailNow()
	}
}
