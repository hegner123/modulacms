package db

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var createTestTable string

func TestCreateDBCopy(t *testing.T) {
	testTable, err := createDbCopy("create_tests.db",false)
	if err != nil {
		logError("failed to create copy of the database, I have to hurry, I'm running out of time!!! ", err)
		t.FailNow()
	}
	createTestTable = testTable
}

func TestCreateUser(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedUser := dbCreateUser(db, ctx, mdb.CreateUserParams{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "systemtest",
		Name:         "systemtest",
		Email:        "test2@modulacmstest.com",
		Hash:         "has",
		Role:         "admin",
	})

	expected := mdb.Users{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "systemtest",
		Name:         "systemtest",
		Email:        "test2@modulacmstest.com",
		Hash:         "has",
		Role:         "admin",
	}

	if reflect.DeepEqual(insertedUser, expected) {
		t.FailNow()
	}
}

func TestCreateAdminRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedAdminRoute := dbCreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       "systemtest",
		AuthorID:     int64(1),
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
		Template:     "page.html",
	})

	expected := mdb.AdminRoutes{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
		Template:     "page.html",
	}

	if reflect.DeepEqual(insertedAdminRoute, expected) {
		t.FailNow()
	}
}

func TestCreateRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedRoute := dbCreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})

	expected := mdb.Routes{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedRoute, expected) {
		t.FailNow()
	}
}

func TestCreateMedia(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedMedia := dbCreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systemtest",
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	})

	expected := mdb.Media{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systemtest",
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	}

	if reflect.DeepEqual(insertedMedia, expected) {
		t.FailNow()
	}
}

func TestCreateDatatype(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	_, err = dbCreateDataType(db, ctx, mdb.CreateDatatypeParams{
		RouteID:      ni(1),
		Label:        "Parent",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		logError("failed to create datatype", err)
	}

	insertedDatatypes, err := dbCreateDataType(db, ctx, mdb.CreateDatatypeParams{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "title",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		logError("failed to create datatype", err)
		t.FailNow()
	}

	expected := mdb.Datatypes{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "title",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     1,
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedDatatypes, expected) {
		t.FailNow()
	}
}

func TestCreateField(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	insertedFields, _ := dbCreateField(db, ctx, mdb.CreateFieldParams{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	expected := mdb.Fields{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedFields, expected) {
		t.FailNow()
	}
}

func TestCreateMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedMediaDimension := dbCreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
        AspectRatio: ns("16:9"),
	})

	expected := mdb.MediaDimensions{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
        AspectRatio: ns("16:9"),
	}

	if reflect.DeepEqual(insertedMediaDimension, expected) {
		t.FailNow()
	}
}

func TestCreateTables(t *testing.T) {
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedTables := dbCreateTable(db, ctx, mdb.Tables{Label: ns("Test")})

	expected := mdb.Tables{
		Label: ns("Test"),
	}

	if reflect.DeepEqual(insertedTables, expected) {
		t.FailNow()
	}
}

func TestCreateToken(t *testing.T) {
	var (
		key         []byte
		tk          *jwt.Token
		signedToken string
	)

	key = generateKey()
	tk = jwt.New(jwt.SigningMethodHS256)
	signedToken, err := tk.SignedString(key)
	if err != nil {
		logError("failed to : ", err)
	}
	db, ctx, err := getDb(Database{src: createTestTable})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	dur, err := time.ParseDuration("24h")
	if err != nil {
		logError("failed to ParseDuration: ", err)
	}
	weeks := time.Now().Add(dur)
	times := timestampS()

	insertedToken := dbCreateToken(db, ctx, mdb.CreateTokenParams{
		UserID:    1,
		IssuedAt:  times,
		ExpiresAt: fmt.Sprint(weeks.Unix()),
		TokenType: "refresh",
		Token:     signedToken,
		Revoked:   nb(false),
	})

	expected := mdb.Tokens{
		UserID:    1,
		IssuedAt:  times,
		ExpiresAt: fmt.Sprint(weeks.Unix()),
		TokenType: "refresh",
		Token:     signedToken,
		Revoked:   nb(false),
	}

	if reflect.DeepEqual(insertedToken, expected) {
		t.FailNow()
	}
}
