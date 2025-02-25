package db

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var getTestTable string

func TestGetDBCopy(t *testing.T) {
	testTable, err := CopyDb("get_tests.db", false)
	if err != nil {
        t.FailNow()
		return
	}

	getTestTable = testTable
}

func TestGetInit(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	file, err := os.ReadFile("./sql/test1.sql")
    if err!=nil {
        return
    }
	s := fmt.Sprintf("%v", file)

	_, err = db.Connection.ExecContext(db.Context, s)
	if err != nil {
		return
	}
}

func TestGetGlobalAdminDatatypeId(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	row, err := GetAdminDatatypeGlobalId(db.Connection, db.Context)
	if err != nil {
		return
	}
	if row.AdminDtID == 0 {
		t.FailNow()
	}
}

func TestGetUser(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(1)

	userRow, err := GetUser(db.Connection, db.Context, id)
	if err != nil {
		return
	}

	expected := mdb.Users{
		UserID:       int64(1),
		DateCreated:  userRow.DateCreated,
		DateModified: userRow.DateModified,
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
		 db := GetDb(Database{DB: getTestTable})

	    id := GetUserId(db,ctx,1)
	}
*/
func TestGetAdminRoute(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	adminRouteRow, err := GetAdminRoute(db.Connection, db.Context, "/admin/login")
	if err != nil {
		return
	}

	expected := mdb.AdminRoutes{
		AdminRouteID: int64(1),
		Author:       "system",
		AuthorID:     1,
		Slug:         "/admin/login",
		Title:        "ModulaCMS",
		Status:       int64(0),
		DateCreated:  adminRouteRow.DateCreated,
		DateModified: adminRouteRow.DateModified,
	}

	if reflect.DeepEqual(adminRouteRow, expected) {
		t.FailNow()
	}
}

func TestGetRoute(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	routeRow, err := GetRoute(db.Connection, db.Context, "/")
	if err != nil {
		return
	}

	expected := mdb.Routes{
		RouteID:      int64(1),
		Author:       "system",
		AuthorID:     1,
		Slug:         "/",
		Title:        "Test",
		Status:       int64(0),
		DateCreated:  routeRow.DateCreated,
		DateModified: routeRow.DateModified,
	}

	if reflect.DeepEqual(routeRow, expected) {
		t.FailNow()
	}
}

func TestGetMedia(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(2)

	mediaRow, err := GetMedia(db.Connection, db.Context, id)
	if err != nil {
		return
	}

	expected := mdb.Media{
		MediaID:            int64(1),
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "system",
		AuthorID:           1,
		DateCreated:        mediaRow.DateCreated,
		DateModified:       mediaRow.DateModified,
		Url:                ns("public/2024/11/test.png1"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	}

	if reflect.DeepEqual(mediaRow, expected) {
		t.FailNow()
	}
}

func TestGetField(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(2)
	fieldRow, err := GetField(db.Connection, db.Context, id)
	if err != nil {
		return
	}

	expected := mdb.Fields{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "title",
		Data:         "Test Field",
		Type:         "text",
		Author:       ns("system"),
		AuthorID:     1,
		DateCreated:  fieldRow.DateCreated,
		DateModified: fieldRow.DateModified,
	}

	if reflect.DeepEqual(fieldRow, expected) {
		t.FailNow()
	}
}

func TestGetMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(2)

	mediaDimensionRow, err := GetMediaDimension(db.Connection, db.Context, id)
	if err != nil {
		return
	}

	expected := mdb.MediaDimensions{
		Label:  ns("Desktop1"),
		Width:  ni(1920),
		Height: ni(1080),
	}

	if reflect.DeepEqual(mediaDimensionRow, expected) {
		t.FailNow()
	}
}

func TestGetTables(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(2)
	tableRow, err := GetTable(db.Connection, db.Context, id)
	if err != nil {
		return
	}

	expected := mdb.Tables{
		Label: ns("Test1"),
	}

	if reflect.DeepEqual(tableRow, expected) {
		t.FailNow()
	}
}
func TestGetTokens(t *testing.T) {
	db := GetDb(Database{Src: getTestTable})

	id := int64(2)
	tokenRow, err := GetToken(db.Connection, db.Context, id)
	if err != nil {
		return
	}
    fmt.Println(tokenRow)


}

