package db

import (
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var updateTestTable string

func TestUpdateDBCopy(t *testing.T) {
	testTable, err := CopyDb("update_tests.db", false)
	if err != nil {
		t.Fatal(err)
		return
	}

	updateTestTable = testTable
}


func TestUpdateAdminRoute(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: updateTestTable})

	params := mdb.UpdateAdminRouteParams{
		Author:       ns("system"),
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateModified: ns(times),
		History:      ns(""),
		Slug_2:       "/test",
	}

	_, err := UpdateAdminRoute(db.Connection, db.Context, params)
	if err != nil {
		t.Fatal(err)
		return

	}
}

func TestUpdateDatatype(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: updateTestTable})

	id := int64(1)
	params := mdb.UpdateDatatypeParams{
		RouteID:      ni64(1),
		Label:        "Parent",
		Type:         "text",
		Author:       "system",
		AuthorID:     1,
        History:      ns(""),
		DateModified: ns(times),
		DatatypeID:   id,
	}

	_, err := UpdateDatatype(db.Connection, db.Context, params)

	if err != nil {
        t.Fatal(err) 
		return
	}

}

func TestUpdateField(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: updateTestTable})

	id := int64(3)
	params := mdb.UpdateFieldParams{
		RouteID:      ni64(1),
		ParentID:     ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "system",
		AuthorID:     1,
		History:      ns(""),
		DateModified: ns(times),
		DateCreated:  ns(times),
		FieldID:      id,
	}

	_, err := UpdateField(db.Connection, db.Context, params)

	if err != nil {
        t.Fatal(err) 
		return
	}
}


func TestUpdateMedia(t *testing.T) {
	db := GetDb(Database{Src: updateTestTable})

	params := mdb.UpdateMediaParams{
		Name:     ns("Best"),
		Author:   "system",
		AuthorID: int64(1),
		MediaID:  int64(2),
	}

	_, err := UpdateMedia(db.Connection, db.Context, params)
	if err != nil {
        t.Fatal(err) 
		return
	}

}

func TestUpdateMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: updateTestTable})

	params := mdb.UpdateMediaDimensionParams{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
	}

	_, err := UpdateMediaDimension(db.Connection, db.Context, params)
	if err != nil {
        t.Fatal(err) 
		return
	}

}
func TestUpdateRoute(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: updateTestTable})

	params := mdb.UpdateRouteParams{
		Author:       "system",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "TEST",
		Status:       0,
		History:      ns(""),
		DateModified: ns(times),
		Slug_2:       "/test",
	}

	_, err := UpdateRoute(db.Connection, db.Context, params)
	if err != nil {
		t.Fatal(err)
		return
	}

}

func TestUpdateTables(t *testing.T) {
	db := GetDb(Database{Src: updateTestTable})

	id := int64(1)
	params := mdb.UpdateTableParams{
		Label: ns("Tested"),
		ID:    id,
	}

	_, err := UpdateTable(db.Connection, db.Context, params)
	if err != nil {
        t.Fatal(err) 
		return
	}
}
func TestUpdateUser(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: updateTestTable})

	id := int64(2)
	params := mdb.UpdateUserParams{
		DateModified: ns(times),
		Name:         "systemupdate",
		Hash:         "has",
		Role:         int64(1),
		UserID:       id,
	}

	_, err := UpdateUser(db.Connection, db.Context, params)
	if err != nil {
		t.Fatal(err)
		return
	}

}
