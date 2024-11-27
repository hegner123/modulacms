package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func TestCreateUser(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedUser := dbCreateUser(db, ctx, mdb.CreateUserParams{
		Datecreated:  ns(times),
		Datemodified: ns(times),
		Username:     ns("system"),
		Name:         ns("system"),
		Email:        ns("test@modulacms.com"),
		Hash:         ns("has"),
		Role:         ns("admin"),
	})

	expected := mdb.User{
		Datecreated:  ns(times),
		Datemodified: ns(times),
		Username:     ns("system"),
		Name:         ns("system"),
		Email:        ns("test@modulacms.com"),
		Hash:         ns("has"),
		Role:         ns("admin"),
	}

	if reflect.DeepEqual(insertedUser, expected) {
		t.FailNow()
	}
}

func TestCreateAdminRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedAdminRoute := dbCreateAdminRoute(db, ctx, mdb.CreateAdminRouteParams{
		Author:       "system",
		Authorid:     "0",
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datecreated:  times,
		Datemodified: times,
		Template:     "page.html",
	})

	expected := mdb.Adminroute{
		Author:       "system",
		Authorid:     "0",
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datecreated:  times,
		Datemodified: times,
		Template:     "page.html",
	}

	if reflect.DeepEqual(insertedAdminRoute, expected) {
		t.FailNow()
	}
}

func TestCreateRoute(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedRoute := dbCreateRoute(db, ctx, mdb.CreateRouteParams{
		Author:       "system",
		Authorid:     "0",
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datecreated:  times,
		Datemodified: times,
		Content:      ns("Test content"),
	})

	expected := mdb.Route{
		Author:       "system",
		Authorid:     "0",
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		Datecreated:  times,
		Datemodified: times,
		Content:      ns("Test content"),
	}

	if reflect.DeepEqual(insertedRoute, expected) {
		t.FailNow()
	}
}

func TestCreateMedia(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedMedia := dbCreateMedia(db, ctx, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "system",
		Authorid:           "0",
		Datecreated:        times,
		Datemodified:       times,
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png"),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png"),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png"),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png"),
	})

	expected := mdb.Media{
		Name:               ns("test.png"),
		Displayname:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "system",
		Authorid:           "0",
		Datecreated:        times,
		Datemodified:       times,
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		Optimizedmobile:    ns("public/2024/11/test-mobile.png"),
		Optimizedtablet:    ns("public/2024/11/test-tablet.png"),
		Optimizeddesktop:   ns("public/2024/11/test-desktop.png"),
		Optimizedultrawide: ns("public/2024/11/test-ultra.png"),
	}

	if reflect.DeepEqual(insertedMedia, expected) {
		t.FailNow()
	}
}

func TestCreateField(t *testing.T) {
	times := timestampS()
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()
	dbCreateField(db, ctx, mdb.CreateFieldParams{
		Routeid:      ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       ns("system"),
		Authorid:     ns("0"),
		Datecreated:  ns(times),
		Datemodified: ns(times),
	})

	insertedField := dbCreateField(db, ctx, mdb.CreateFieldParams{
		Routeid:      ni(1),
		Parentid:     ni(1),
		Label:        "title",
		Data:         "Test Field",
		Type:         "text",
		Struct:       ns("text"),
		Author:       ns("system"),
		Authorid:     ns("0"),
		Datecreated:  ns(times),
		Datemodified: ns(times),
	})

	expected := mdb.Field{
		Routeid:      ni(1),
		Parentid:     ni(1),
		Label:        "title",
		Data:         "Test Field",
		Type:         "text",
		Struct:       ns("text"),
		Author:       ns("system"),
		Authorid:     ns("0"),
		Datecreated:  ns(times),
		Datemodified: ns(times),
	}

	if reflect.DeepEqual(insertedField, expected) {
		t.FailNow()
	}
}

func TestCreateMediaDimension(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
	if err != nil {
		logError("failed to connect or create database", err)
	}
	defer db.Close()

	insertedMediaDimension := dbCreateMediaDimension(db, ctx, mdb.CreateMediaDimensionParams{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
	})

	expected := mdb.MediaDimension{
		Label:  ns("Desktop"),
		Width:  ni(1920),
		Height: ni(1080),
	}

	if reflect.DeepEqual(insertedMediaDimension, expected) {
		t.FailNow()
	}
}

func TestCreateTables(t *testing.T) {
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
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
	db, ctx, err := getDb(Database{DB: "modula_test.db"})
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

	expected := mdb.Token{
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
